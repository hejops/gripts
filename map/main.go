// https://www.ardanlabs.com/blog/2021/11/gis-in-go.html
// https://github.com/353words/track/blob/master/track.go

// csv parsing, time parsing, template embedding

package main

import (
	_ "embed" // silly go
	"encoding/csv"
	"html/template"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jszwec/csvutil"
)

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

type Row struct {
	// to deserialise into a struct, fields must:
	// 1. be Uppercase
	// 2. be annotated (according to header values in file)

	// https://go.dev/ref/spec#Struct_types

	Time   time.Time `csv:"time"`
	Lat    float64   `csv:"lat"`
	Lng    float64   `csv:"lng"`
	Height float64   `csv:"height"`
}

type Composer struct {
	FirstName string  `csv:"firstname"`
	LastName  string  `csv:"lastname"`
	Born      int     `csv:"born"`
	Died      int     `csv:"died"`
	Place     string  `csv:"place"`
	Lat       float64 `csv:"lat"`
	Long      float64 `csv:"long"`
}

func main() {
	// b, err := os.ReadFile("./map.csv")
	b, err := os.ReadFile("./composers.csv")
	if err != nil {
		log.Fatal(err)
	}

	// bytes -> string -> string reader -> csv reader -> csv decoder
	dec, err := csvutil.NewDecoder(csv.NewReader(strings.NewReader(string(b))))
	if err != nil {
		log.Fatal(err)
	}

	// failure to parse will raise err, and result in null values
	// dec.Register(parseTime) // deprecated
	dec.WithUnmarshalers(csvutil.UnmarshalFunc(parseTime))

	// string -> []Row
	// note that we don't actually iterate through a []string, rather we
	// Decode an entire csv object until we reach EOF

	var rows []Composer
	for {

		// i never liked this "direct the contents of some operation to
		// an initialised var" idiom (Rust has it too, in some read
		// methods)
		// https://doc.rust-lang.org/std/fs/struct.File.html#method.open
		var row Composer
		err := dec.Decode(&row)

		// fmt.Println(row)
		if err == io.EOF {
			break
		}
		rows = append(rows, row)
	}

	// we skip the resampling part (for now)

	// inject state (from csv) into the template
	data := map[string]interface{}{
		"rows": rows,
	}
	if err := mapTemplate.Execute(os.Stdout, data); err != nil {
		log.Fatal(err)
	}
}
