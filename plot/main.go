// https://github.com/gonum/plot/wiki/Example-plots

package main

import (
	"fmt"
	"io"
	"os"

	"encoding/json/jsontext"
	"encoding/json/v2"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func trunc(s plotter.XYs) plotter.XYs {
	// an uninitialised XY (in a XYs) results in bogus plots. furthermore,
	// XYs is fixed length (not growable), but we cannot know len upfront.
	var n int
	for i, pt := range s {
		if pt.Y == 0 {
			break
		}
		n = i
	}
	n++
	out := make(plotter.XYs, n)
	for i := range n {
		out[i] = s[i]
	}
	// fmt.Println(out)
	return out
}

func spread(m map[string]plotter.XYs) []any {
	var s []any
	for k, v := range m {
		s = append(s, k, trunc(v))
	}
	return s
}

func main() {
	// input must be jsonl
	// {"foo": 1, "bar": 2}
	// {"foo": 2, "bar": 4}

	f, _ := os.Open(os.Getenv("HOME") + "/foo.jsonl")
	d := jsontext.NewDecoder(f)

	rows := make(map[string]plotter.XYs)

	for i := 0; ; i++ {
		var r map[string]int
		if err := json.UnmarshalDecode(d, &r); err == io.EOF {
			break // done decoding file
		} else if err != nil {
			panic(err)
		}

		for k, v := range r {
			if len(rows[k]) == 0 {
				rows[k] = make(plotter.XYs, 12)
			}
			rows[k][i].X = float64(i) + 1
			rows[k][i].Y = float64(v)
		}
	}

	fmt.Println(rows)

	p := plot.New()

	// to get different colours per line, all string+XYs pairs must be
	// passed in a single AddLinePoints call
	if err := plotutil.AddLinePoints(p, spread(rows)...); err != nil {
		panic(err)
	}

	// p.X.AutoRescale = true // doesn't actually prevent bogus plots
	// p.Y.AutoRescale = true

	// TODO: x axis all major ticks, interval 1

	// Save the plot to a PNG file.
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "points.png"); err != nil {
		panic(err)
	}

	fmt.Println("ok")
}
