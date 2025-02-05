package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const NTFS_ILLEGAL = `<>:"/\|?*`

var (
	// bufSize   = flag.Int("b", 4096, "buffer size") // only for copy
	SRC_BASE  string
	DEST_BASE string

	images = map[string]any{
		"Folder.jpg": nil,
		"folder.jpg": nil,
		"Cover.jpg":  nil,
		"cover.jpg":  nil,
		"cover.png":  nil,
	}
)

func init() {
	SRC_BASE = os.Getenv("SLSK")
	if SRC_BASE == "" {
		panic(1)
	}
	SRC_BASE += "/complete"

	DEST_BASE = os.Getenv("MU")
	if DEST_BASE == "" {
		panic(1)
	}
}

// func copy(src string, dest string) error { // {{{
// 	// https://github.com/torvalds/linux/blob/master/fs/ext4/ext4.h#L2292
// 	// ext4: 255 bytes
// 	// ntfs: 255 utf-16
// 	if len(dest) > 255 {
// 		panic("filename truncation not impl yet")
// 	}
//
// 	// https://opensource.com/article/18/6/copying-files-go
// 	srcStat, err := os.Stat(src)
// 	if err != nil {
// 		return err
// 	} else if !srcStat.Mode().IsRegular() {
// 		return fmt.Errorf("%s is not a regular file", src)
// 	}
//
// 	srcFile, err := os.Open(src)
// 	if err != nil {
// 		return err
// 	}
// 	defer srcFile.Close()
//
// 	if destStat, err := os.Stat(dest); err == nil {
// 		if destStat.IsDir() {
// 			return errors.New("dest cannot be a dir")
// 		}
// 		os.Remove(dest)
// 	}
//
// 	_ = os.MkdirAll(filepath.Dir(dest), 0700)
// 	destFile, err := os.Create(dest)
// 	if err != nil {
// 		return err
// 	}
// 	defer destFile.Close()
//
// 	buf := make([]byte, *bufSize)
// 	for {
// 		n, err := srcFile.Read(buf)
// 		if err != nil && err != io.EOF {
// 			return err
// 		}
// 		if n == 0 {
// 			break
// 		}
//
// 		if _, err := destFile.Write(buf[:n]); err != nil {
// 			return err
// 		}
// 	}
// 	return err
// } // }}}

func (t *Tags) Move() (string, error) {
	dest := t.Destination()
	_ = os.MkdirAll(filepath.Dir(dest), 0700)
	err := os.Rename(t.Path, dest)
	if err != nil {
		return "", err
	}
	return dest, nil
}

// func (t *Tags) Link() error {

// func Link(src string, dest string) error {
// 	_ = os.MkdirAll(filepath.Dir(dest), 0700)
// 	return os.Symlink(src, dest)
// }

func moveDir(d string) { // {{{
	// when attempting to move a large number of directories, there is no
	// reliable way to tell whether all files are present; it is fairly
	// common for some files to be missing. directories must explicitly be
	// marked as done by the tagger
	var processed bool

	var files []string
	_ = filepath.WalkDir(d, func(
		path string,
		d fs.DirEntry,
		err error,
	) error {
		if !d.IsDir() && filepath.Ext(path) == ".mp3" {
			files = append(files, path)
		} else if d.Name() == "done" {
			processed = true
		}
		return nil
	})

	if !processed {
		fmt.Println("not processed:", d)
		return
	}

	// if len(files) > 99 {
	// 	panic("zero padding for 100 tracks not impl yet")
	// }

	var image string
	var destDir string

	var allTags []Tags
	var lastTrack uint16
	var numFiles int
	artists := make(map[string]any)

	for _, p := range files {

		tags, err := NewTags(p)
		if err != nil {
			if _, ok := images[filepath.Base(p)]; ok {
				image = p
			}
			// fmt.Println(err.Error(), p) // usually "invalid file"
			continue
		}

		allTags = append(allTags, tags)
		lastTrack = max(lastTrack, tags.TrackNumber)
		numFiles++
		artists[tags.Artist] = nil

		if destDir == "" {
			destDir = filepath.Dir(tags.Destination())
		}
	}

	if len(allTags) == 0 {
		fmt.Fprintln(os.Stderr, "nothing to move:", d)
		return
	} else if numFiles != int(lastTrack) {
		fmt.Fprintln(os.Stderr, "not enough files:", d)
		return
	}

	va := len(artists) > 1
	type Link struct {
		oldname string
		newname string
	}
	var links []Link

	// fmt.Println(allTags[0].rawTags)

	for _, tags := range allTags {
		fmt.Println(tags.Path)
		if *MOVE {
			dest, err := tags.Move()
			if err != nil {
				panic(err)
			}

			if va {
				for art := range artists {
					// fmt.Println(art, tags.Artist, art == tags.Artist)
					if art != tags.Artist {
						oldname := strings.Replace(dest, DEST_BASE, "../..", 1)

						// must make a copy, otherwise only one 'overwrite' will succeed
						copied := tags
						copied.Artist = art
						newname := copied.Destination()

						// link target must exist
						// before a symlink can be
						// made, so we cannot symlink
						// in this loop yet!
						links = append(links, Link{oldname: oldname, newname: newname})
					}
				}
			}
		}
		// fmt.Println(tags.Destination())
	}

	for _, l := range links {
		_ = os.Symlink(l.oldname, l.newname)
		// err := os.Symlink(l.oldname, l.newname)
		// if err != nil {
		// 	fmt.Println(err)
		// }
		fmt.Println(l.newname, "->", l.oldname)
	}

	if *MOVE && image != "" {
		// err := copy(image, filepath.Join(dest, "folder.jpg"))
		err := os.Rename(image, filepath.Join(destDir, "folder.jpg"))
		if err != nil {
			panic(err)
		}
	}

	fmt.Println(destDir) // output should then be piped to wherever

	// p := "/tmp/fjksdlfjdask"
	// c := exec.Command("fix_tags", p)
	// out, err := c.Output()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
} // }}}
