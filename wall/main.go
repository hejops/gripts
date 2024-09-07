// Simple wallpaper setter

package main

import (
	"log"
	"math/rand/v2"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// note: this program has the same name as /usr/bin/wall

func main() {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "wallpaper")

	files := []string{}
	err := filepath.Walk(dir, func(
		// Walk is annoying to use since it requires this specific type
		// sig (without explicitly saying it -- see example), but it
		// does provide fullpath (like os.scandir), and is better
		// generalisable to arbitrary directory structures
		path string, info os.FileInfo, e error,
	) error {
		if strings.Contains(path, ".git") {
			return nil
		}

		if !info.IsDir() {
			files = append(files, path)
		}
		return e
	})
	if err != nil {
		log.Fatal(err)
	}

	var file string
	for _, i := range rand.Perm(len(files)) {
		file = files[i]
		ext := filepath.Ext(file)
		if mime.TypeByExtension(ext)[:5] != "image" {
			continue
		}
		// i could select 2 images, but i don't use 2 monitors any more
		if exec.Command(
			"feh",
			"--no-fehbg",
			"--bg-fill",
			// files[rand.N(len(files)-1)],
			file,
		).Run() != nil {
			log.Fatal(err)
		}
		return
	}

	panic("no file found")
}
