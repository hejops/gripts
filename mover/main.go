package main

import (
	"flag"
	"io/fs"
	"path/filepath"
)

var MOVE = flag.Bool("move", false, "buffer size") // will be removed eventually

func main() {
	flag.Parse()

	// cp complete complete_
	// time go run *.go

	// 2030 MB = 0m0.700s (go)

	// for dirs with automatically detected errors, just skip
	//
	// TODO: show list of dirs in tea, manually ignore (for now, until i impl tag writing in go)
	//
	// move the rest

	var dirs []string
	_ = filepath.WalkDir(SRC_BASE, func(
		path string,
		d fs.DirEntry,
		err error,
	) error {
		if d.IsDir() && filepath.Dir(path) == SRC_BASE {
			dirs = append(dirs, path)
		}
		return nil
	})

	for _, d := range dirs {
		// fmt.Println(d)
		moveDir(d)
	}
}
