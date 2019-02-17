package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	base := "data/GeoLite2-City-Locations"
	locale := []string{"de", "es", "pt-BR", "en", "fr", "ru", "ja", "zh-CN"}

	outDir := "js"
	err := os.MkdirAll(outDir, 0755)
	if err != nil {
		panic(err)
	}

	for _, loc := range locale {
		input := fmt.Sprintf("%s-%s.csv", base, loc)
		geoList := GenCountryList(input, "")

		err = genJS(geoList, outDir, loc)
		if err != nil {
			panic(err)
		}
	}
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
	err := genJSObj(cs, filepath.Join(outDir, "map.js"))
	if err != nil {
		return err
	}

	// js array of names
	err = genJSArr(cs, filepath.Join(outDir, "list.js"))
	if err != nil {
		return err
	}

	return nil
}

func genJSObj(geoList []Country, out string) error {
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	size := len(geoList)
	f.WriteString("export default {\n")
	for i, c := range geoList {
		if i < size-1 {
			f.WriteString(fmt.Sprintf("  %s: '%s',\n", c.Code, c.Name))
		} else {
			f.WriteString(fmt.Sprintf("  %s: '%s'\n", c.Code, c.Name))
		}
	}
	f.WriteString("}")
	return nil
}

func genJSArr(geoList []Country, out string) error {
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	sort.Sort(ByName(geoList))

	size := len(geoList)
	f.WriteString("export default [\n")
	for i, c := range geoList {
		if i < size-1 {
			f.WriteString(fmt.Sprintf("  '%s',\n", c.Name))
		} else {
			f.WriteString(fmt.Sprintf("  '%s'\n", c.Name))
		}
	}
	f.WriteString("]")
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

	// unique code and country name
	countDict := make(map[string]Country)
	for i, l := range lines {
		code, name := l[4], l[5]
		if i == 0 || code == "" || name == "" {
			continue
		}
		countDict[code] = Country{code, name}
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
