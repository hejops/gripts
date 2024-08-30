package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

const (
	INDEX      = "https://index.golang.org/index"
	PKG_SEARCH = "https://pkg.go.dev/search?limit=25&m=package&q="
)

var GO_PKG_BASE = os.ExpandEnv("$HOME/go/pkg/mod")

type IndexPackage struct {
	Path      string
	Timestamp string
	Version   string
}

type SearchPackage struct {
	Path      string
	Synopsis  string
	Downloads int // TODO:
}

func isVer(name string) bool {
	return name[0] == 'v' && unicode.IsNumber(rune(name[1]))
}

// Uses pkg.go.dev, which generally returns more results
func findPackage(name string) []SearchPackage {
	fmt.Println(PKG_SEARCH + name)
	resp, err := http.Get(PKG_SEARCH + name)
	if err != nil {
		panic(err)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}

	// fmt.Println(doc.First().Html())

	pkgs := []SearchPackage{}
	doc.Find("span").Each(func(i int, sel *goquery.Selection) {
		if c, _ := sel.Attr("class"); c == "SearchSnippet-header-path" {
			pkg := sel.Text()
			pkg = pkg[1 : len(pkg)-1] // strip parens
			base := path.Base(pkg)
			if base == name || isVer(base) {
				// yeah...
				syn := sel.Parent().Parent().Parent().Parent().Find("p").Text()
				syn = strings.TrimSpace(syn)
				pkgs = append(
					pkgs,
					SearchPackage{Path: pkg, Synopsis: syn},
				)
			}
		}
	})

	return pkgs
}

func findIndexPackage(name string) []IndexPackage {
	resp, err := http.Get(INDEX)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// can't use the standard json.Unmarshal because of jsonl

	var pkgs []IndexPackage

	// https://stackoverflow.com/a/34388102
	d := json.NewDecoder(resp.Body)
	for {
		var pkg IndexPackage
		if err := d.Decode(&pkg); err == io.EOF {
			break // done decoding file
		} else if err != nil {
			panic(err)
		}
		if path.Base(pkg.Path) == name {
			pkgs = append(pkgs, pkg)
		}
	}

	return pkgs
}
