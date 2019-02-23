package main

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const sourceURL = "https://geolite.maxmind.com/download/geoip/database/GeoLite2-City-CSV.zip"
const csvPrefix = "GeoLite2-City-Locations"

func main() {
	dataDir, err := prepareData("data")
	if err != nil {
		panic(err)
	}

	locale := []string{"de", "es", "pt-BR", "en", "fr", "ru", "ja", "zh-CN"}

	outDir := "js"
	err = os.MkdirAll(outDir, 0755)
	if err != nil {
		panic(err)
	}

	log.Printf("Parsing %s\n", dataDir)
	var geoList []Country
	for _, loc := range locale {
		input := fmt.Sprintf("%s-%s.csv", csvPrefix, loc)
		geoList = GenCountryList(filepath.Join(dataDir, input), "")

		err = genJS(geoList, outDir, loc)
		if err != nil {
			panic(err)
		}
	}

	// js array of codes
	err = genJSArrCodes(geoList, filepath.Join(outDir, "codes.js"))
	if err != nil {
		panic(err)
	}

	log.Println("Done")
}

func prepareData(dir string) (string, error) {
	_, err := os.Stat(dir)
	if err == nil {
		// @TODO: find the latest GeoLite2-City-CSV dir
		return "./data/GeoLite2-City-CSV_20190219", nil
	}
	err = os.Mkdir(dir, 0755)
	if err != nil {
		return "", err
	}

	log.Printf("Downloading %s...\n", sourceURL)
	z := filepath.Join(dir, "csv.zip")
	err = DownloadFile(z, sourceURL)
	if err != nil {
		return "", err
	}

	log.Println("Extracting...")
	dataDir, err := Unzip(z, dir)
	if err != nil {
		return "", err
	}
	err = os.Remove(z)
	if err != nil {
		return "", err
	}

	return dataDir, nil
}

// DownloadFile downloads from url and save to filepath.
func DownloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// Unzip unzips the zip file and return the name of extracted dir.
func Unzip(src, dest string) (string, error) {
	var extractedDir string
	r, err := zip.OpenReader(src)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()

		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
		} else {
			var fdir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				fdir = fpath[:lastIndex]
			}

			err = os.MkdirAll(fdir, 0755)
			if err != nil {
				log.Fatal(err)
				return "", err
			}
			if extractedDir == "" {
				extractedDir = fdir
			}
			f, err := os.OpenFile(
				fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return "", err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return "", err
			}
		}
	}
	return extractedDir, nil
}

func genJS(cs []Country, dir, locale string) error {
	shortLocale := strings.ToLower(strings.Replace(locale, "-", "", -1))
	outDir := filepath.Join(dir, shortLocale)
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		err := os.Mkdir(outDir, 0755)
		if err != nil {
			return err
		}
	}

	// js object: code to name
	err := genJSObjCodeToName(cs, filepath.Join(outDir, "map-code-name.js"))
	if err != nil {
		return err
	}

	// js object: name to code
	err = genJSObjNameToCode(cs, filepath.Join(outDir, "map-name-code.js"))
	if err != nil {
		return err
	}

	// js array of names
	err = genJSArrNames(cs, filepath.Join(outDir, "names.js"))
	if err != nil {
		return err
	}

	return nil
}

func genJSObjCodeToName(geoList []Country, out string) error {
	log.Printf("Generating %s...\n", out)
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	size := len(geoList)
	f.WriteString("export default () => ({\n")
	for i, c := range geoList {
		if i < size-1 {
			f.WriteString(fmt.Sprintf("  %s: '%s',\n", c.Code, c.Name))
		} else {
			f.WriteString(fmt.Sprintf("  %s: '%s'\n", c.Code, c.Name))
		}
	}
	f.WriteString("})\n")
	return nil
}

func genJSObjNameToCode(geoList []Country, out string) error {
	log.Printf("Generating %s...\n", out)
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	sort.Sort(ByName(geoList))

	size := len(geoList)
	f.WriteString("export default () => ({\n")
	for i, c := range geoList {
		if i < size-1 {
			f.WriteString(fmt.Sprintf("  '%s': '%s',\n", c.Name, c.Code))
		} else {
			f.WriteString(fmt.Sprintf("  '%s': '%s'\n", c.Name, c.Code))
		}
	}
	f.WriteString("})\n")
	return nil
}

func genJSArrNames(geoList []Country, out string) error {
	log.Printf("Generating %s...\n", out)
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	sort.Sort(ByName(geoList))

	size := len(geoList)
	f.WriteString("export default () => ([\n")
	for i, c := range geoList {
		if i < size-1 {
			f.WriteString(fmt.Sprintf("  '%s',\n", c.Name))
		} else {
			f.WriteString(fmt.Sprintf("  '%s'\n", c.Name))
		}
	}
	f.WriteString("])\n")
	return nil
}

func genJSArrCodes(geoList []Country, out string) error {
	log.Printf("Generating %s...\n", out)
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	sort.Sort(ByCode(geoList))

	size := len(geoList)
	f.WriteString("export default () => ([\n")
	for i, c := range geoList {
		if i < size-1 {
			f.WriteString(fmt.Sprintf("  '%s',\n", c.Code))
		} else {
			f.WriteString(fmt.Sprintf("  '%s'\n", c.Code))
		}
	}
	f.WriteString("])\n")
	return nil
}

// GenCountryList generates a sorted list of unique code and country name.
func GenCountryList(csvPath, outPath string) []Country {
	// read
	fRead, err := os.Open(csvPath)
	if err != nil {
		panic(err)
	}
	defer fRead.Close()

	reader := csv.NewReader(fRead)
	lines, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	// for removing accents
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

	// unique code and country name
	countDict := make(map[string]Country)
	for i, l := range lines {
		code, name := l[4], l[5]
		if i == 0 || code == "" || name == "" {
			continue
		}
		n, _, _ := transform.String(t, name)
		countDict[code] = Country{code, n}
	}

	// sort
	var countries []Country
	for _, v := range countDict {
		countries = append(countries, v)
	}
	sort.Sort(ByCode(countries))

	// write
	if outPath != "" {
		fWrite, err := os.Create(outPath)
		if err != nil {
			panic(err)
		}
		defer fWrite.Close()

		writer := csv.NewWriter(fWrite)
		defer writer.Flush()

		for _, v := range countries {
			r := []string{v.Code, v.Name}
			err = writer.Write(r)
			if err != nil {
				panic(err)
			}
		}
	}

	return countries
}

// Country represents the iso code and country name.
type Country struct {
	Code string
	Name string
}

// ByCode sorts country by code.
type ByCode []Country

func (a ByCode) Len() int {
	return len(a)
}

func (a ByCode) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByCode) Less(i, j int) bool {
	return a[i].Code < a[j].Code
}

// ByName sorts country by name.
type ByName []Country

func (a ByName) Len() int {
	return len(a)
}

func (a ByName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByName) Less(i, j int) bool {
	return a[i].Name < a[j].Name
}
