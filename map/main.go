// https://www.ardanlabs.com/blog/2021/11/gis-in-go.html

package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jszwec/csvutil"
)

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

	// failure to parse will raise err, and result in null values
	Time   time.Time `csv:"time"`
	Lat    float64   `csv:"lat"`
	Lng    float64   `csv:"lng"`
	Height float64   `csv:"height"`
}

func main() {
	b, err := os.ReadFile("./map.csv")
	if err != nil {
		log.Fatal(err)
	}

	// bytes -> string -> string reader -> csv reader -> csv decoder
	dec, err := csvutil.NewDecoder(csv.NewReader(strings.NewReader(string(b))))
	if err != nil {
		log.Fatal(err)
	}

	// dec.Register(parseTime) // deprecated
	dec.WithUnmarshalers(csvutil.UnmarshalFunc(parseTime))

	// string -> []Row
	// note that we don't actually iterate through a []string, rather we
	// Decode an entire csv object until we reach EOF

	var rows []Row
	for {

		// i never liked this "direct the contents of some operation to
		// an initialised var" idiom (Rust has it too, in some read
		// methods)
		// https://doc.rust-lang.org/std/fs/struct.File.html#method.open
		var row Row
		err := dec.Decode(&row)

		// fmt.Println(row)
		if err == io.EOF {
			break
		}
		rows = append(rows, row)
	}

	fmt.Println(rows)
	fmt.Println(len(rows))
}
