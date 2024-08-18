package main

import (
	"log"
	"math/rand/v2"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

func main() {
	u, _ := user.Current()
	dir := filepath.Join(u.HomeDir, "wallpaper")

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

	// i could select 2 images, but i don't use 2 monitors any more
	if exec.Command(
		"feh", "--no-fehbg", "--bg-fill", files[rand.N(len(files))-1],
	).Run() != nil {
		log.Fatal(err)
	}
}
