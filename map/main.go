// https://www.ardanlabs.com/blog/2021/11/gis-in-go.html
// https://github.com/353words/track/blob/master/track.go

// csv parsing, time parsing, template embedding

package main

import (
	_ "embed" // silly go
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	gowikidata "github.com/Navid2zp/go-wikidata"
	"github.com/jszwec/csvutil"
)

// TODO: refactor to composer.go

type Composer3 struct {
	// wikidata (digits, prefixed with Q)
	ID    string
	Name  string
	Place string
}

func newComposer(name string) Composer3 {
	id := getComposerID(name)
	c := Composer3{
		ID:   id,
		Name: name,
	}
	c.setLocation()
	return c
}

// name -> born/died/place(id) -> place(name)

// https://www.thedataschool.co.uk/rachel-costa/scraping-wikipedia/
// https://www.wikidata.org/w/api.php?action=wbgetentities&ids=Q7070
// https://www.wikidata.org/w/api.php?action=wbgetentities&sites=enwiki&titles=Johann_Sebastian_Bach
// https://www.wikidata.org/wiki/Wikidata:List_of_properties

// ideally, i would've preferred to use a simpler API like openopus, but its
// lack of geographical information is a dealbreaker
// https://github.com/shanewilliams29/composer-explorer-vue/blob/main/data/tables/composer_list.csv -- has years, but no place -- s/\v"([^"]+)";/\1 string `csv:"\1"`\r/g
// https://github.com/openopus-org/openopus_api/blob/master/USAGE.md#list-composers-by-first-letter -- same (json), also has epoch; consider using sql/duckdb

// TODO: geolocation (place -> lat/long)

func getComposerID(name string) string {
	name = url.PathEscape(name)
	res, err := gowikidata.NewSearch(name, "en")
	if err != nil {
		log.Fatal(err)
	}
	ent, err := res.Get()
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range ent.SearchResult {
		return r.ID
	}
	return ""
}

// func setLocation(id string) string {
func (c *Composer3) setLocation() {
	req, err := gowikidata.NewGetClaims(c.ID, "")
	if err != nil {
		log.Fatal(err)
	}

	// Date of birth is Property:P569
	// Place of birth is Property:P19
	// Date of death is Property:570
	req.SetProperty("P19") // TODO: as enum/map

	obj, err := req.Get()
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range *obj {
		// lmao
		loc := getPlaceName(v[0].MainSnak.DataValue.Value.ValueFields.ID)
		c.Place = loc
		return // loc
	}
	// return ""
}

func getPlaceName(id string) string {
	req, err := gowikidata.NewGetEntities([]string{id})
	if err != nil {
		log.Fatal(err)
	}
	req.SetSites([]string{"enwiki"})
	res, err := req.Get()
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range *res {
		return v.GetLabel("en")
	}
	return ""
}

// embed must be done at top-level!
var (
	// load contents of template.html into mapHTML
	//go:embed template.html
	mapHTML string
	// then generate a template from the string
	// this does not need to be at top-level, but it is clearer
	mapTemplate = template.Must(template.New("map").Parse(mapHTML))
)

// parse received bytes and pass them to t via pointer. this type signature is
// required by csvutil.UnmarshalFunc
func parseTime(bytes []byte, t *time.Time) error {
	var err error
	// 2015-08-20 03:48:30.669
	// "01/02 03:04:05PM '06 -0700"
	*t, err = time.Parse("2006-01-02 15:04:05.000", string(bytes))
	return err
}

type Row struct { // {{{
	// to deserialise into a struct, fields must:
	// 1. be Uppercase
	// 2. be annotated (according to header values in file)
	//
	// field order does not affect parsing, but it does affect display

	// https://go.dev/ref/spec#Struct_types

	Time   time.Time `csv:"time"`
	Lat    float64   `csv:"lat"`
	Lng    float64   `csv:"lng"`
	Height float64   `csv:"height"`
} // }}}

// target schema
type Composer struct {
	FirstName string  `csv:"firstname"`
	LastName  string  `csv:"lastname"`
	Born      int     `csv:"born"`
	Died      int     `csv:"died"`
	Place     string  `csv:"place"`
	Lat       float64 `csv:"lat"`
	Long      float64 `csv:"long"`
}

// schema from composer-explorer-vue
type Composer2 struct { // {{{
	// Id string `csv:"id"`
	// Source    string `csv:"source"`
	NameShort string `csv:"name_short"`
	NameFull  string `csv:"name_full"`
	Born      string `csv:"born"`
	Died      string `csv:"died"`
	Country   string `csv:"nationality"`
	// Name_norm  string `csv:"name_norm"`
	// linkname    string `csv:"linkname"`
	// Region      string `csv:"region"`
	// description   string `csv:"description"`
	// image         string `csv:"image"`
	// imgfull       string `csv:"imgfull"`
	// pageurl       string `csv:"pageurl"`
	// wordcount     string `csv:"wordcount"`
	// introduction  string `csv:"introduction"`
	// rank          string `csv:"rank"`
	// spotify       string `csv:"spotify"`
	// clicks        string `csv:"clicks"`
	// catalogued    string `csv:"catalogued"`
	// female        string `csv:"female"`
	// general       string `csv:"general"`
	// view          string `csv:"view"`
	// preview_music string `csv:"preview_music"`
	// tier          string `csv:"tier"`
} // }}}

func main() {
	c := newComposer("Johann Sebastian Bach")
	fmt.Println(c.Place)
	return

	// b, err := os.ReadFile("./map.csv")
	// b, err := os.ReadFile("./composers.csv")
	b, err := os.ReadFile("./composer_list.csv")
	if err != nil {
		log.Fatal(err)
	}

	// bytes -> string -> string reader -> csv reader -> csv decoder
	reader := csv.NewReader(strings.NewReader(string(b)))
	reader.Comma = ';'
	dec, err := csvutil.NewDecoder(reader)
	if err != nil {
		log.Fatal(err)
	}

	// failure to parse will raise err, and result in null values
	// dec.Register(parseTime) // deprecated
	dec.WithUnmarshalers(csvutil.UnmarshalFunc(parseTime))

	// string -> []Row
	// note that we don't actually iterate through a []string, rather we
	// Decode an entire csv object until we reach EOF

	var rows []Composer2
	for {

		// i never liked this "direct the contents of some operation to
		// an initialised var" idiom (Rust has it too, in some read
		// methods)
		// https://doc.rust-lang.org/std/fs/struct.File.html#method.open
		var row Composer2
		err := dec.Decode(&row)

		// fmt.Println(row)
		if err == io.EOF {
			break
		}
		rows = append(rows, row)

		fmt.Println(row)

	}

	// we skip the resampling part (for now)

	// // inject state (from csv) into the template
	// data := map[string]interface{}{
	// 	"rows": rows,
	// }
	// // TODO: write to file (not stdout)
	// if err := mapTemplate.Execute(os.Stdout, data); err != nil {
	// 	log.Fatal(err)
	// }
}
