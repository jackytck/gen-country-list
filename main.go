package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
)

func main() {
	locale := []string{"de", "es", "pt-BR", "en", "fr", "ru", "ja", "zh-CN"}
	base := "data/GeoLite2-City-Locations"

	for _, l := range locale {
		input := fmt.Sprintf("%s-%s.csv", base, l)
		output := fmt.Sprintf("js/country-%s.js", l)
		genJS(input, output)
	}
}

func genJS(csvIn, jsOut string) {
	geoList := GenCountryList(csvIn, "")

	f, err := os.Create(jsOut)
	if err != nil {
		panic(err)
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
