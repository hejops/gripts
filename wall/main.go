package main

import (
	"log"
	"math/rand/v2"
	"os"
	"os/exec"
	"os/user"
	"path"
)

func main() {
	u, _ := user.Current()
	dir := u.HomeDir + "/wallpaper"

	// files := []string{}
	// err := filepath.Walk(dir, func(
	// 	// Walk is annoying to use since it requires this specific type
	// 	// sig (without explicitly saying it -- see example), but it does
	// 	// provide fullpath (like os.scandir).
	// 	//
	// 	// the problem: it is recursive, so we can't stop at depth 1
	// 	path string, info os.FileInfo, err error,
	// ) error {
	// 	if !info.IsDir() {
	// 		files = append(files, path)
	// 	}
	// 	return err
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// entries, err := ioutil.ReadDir(dir) // deprecated in favour of os.ReadDir
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	files := []string{}
	for _, e := range entries {
		if !e.IsDir() {
			// just like Python os.listdir, ReadDir only returns
			// basenames
			files = append(files, path.Join(dir, e.Name()))
		}
	}
	rand.Shuffle(len(files), func(i, j int) {
		files[i], files[j] = files[j], files[i]
	})
	wall := files[0]

	if err := exec.Command(
		"feh", "--no-fehbg", "--bg-fill", wall,
	).Run(); err != nil {
		log.Fatal(err)
	}
}
