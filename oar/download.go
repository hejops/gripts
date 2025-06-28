//

package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cheggaaa/pb/v3"
	"github.com/wader/goutubedl"
)

type Downloadable interface {
	// download(bar *pb.ProgressBar)

	// url() string
	// title() string
	Download(bar *pb.ProgressBar) error

	// unlike in Rust, default impls of an interface must be declared
	// outside the interface
}

var (
	_ Downloadable = (*BandcampRelease)(nil)
	_ Downloadable = (*YoutubeVideo)(nil)
)

func (b *BandcampRelease) Download(bar *pb.ProgressBar) error {
	return download(b.Title, b.Url, bar)
}

func (y *YoutubeVideo) Download(bar *pb.ProgressBar) error {
	return download(y.Title, y.Url, bar)
}

func download(title string, url string, bar *pb.ProgressBar) error { // concrete form

	// to avoid UI glitches, nothing should ever be printed in this func

	_ = os.Mkdir(Cfg.Dest, os.ModePerm)

	ext := ".mp3" // yt's default audio format is mp4 dash, but mpv doesn't care

	fname := strings.ReplaceAll(title+ext, "/", "-")
	path := filepath.Join(Cfg.Dest, fname)

	if _, err := os.Stat(path); err == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	maxMinutes := 60
	meta, _ := goutubedl.New(context.TODO(), url, goutubedl.Options{})
	if dur := int(meta.Info.Duration) / maxMinutes; dur > 60 {
		return fmt.Errorf("skipping long video (%d min)", dur)
	}

	res, err := goutubedl.Download(
		ctx,
		url,
		goutubedl.Options{
			// only relevant for bc
			PlaylistStart: 1, // 0 or 1 indexed?
			PlaylistEnd:   1,
		},
		"bestaudio", // 219 files: 4.89 G
		// "bestaudio[abr<=128]", // 2.73 G
		// abr<=128 = 50-60% of default quality
	)
	if err != nil {
		// i suspect 403s are due to abr<=128
		return err
	}
	defer res.Close()

	// _ = os.Remove(fname)
	file, err := os.Create(path) // replaces existing file
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// see also (mpb): https://github.com/FantomeBeignet/y2storj/blob/7224cad959e95d9cedcb5f07ec7049764b3ab0ab/y2storj.go#L95
	// https://github.com/vbauerster/mpb#rendering-multiple-bars

	var dest io.Writer
	if bar == nil {
		dest = file
	} else {
		dest = bar.NewProxyWriter(file)
		// bar.Start()
	}

	if _, err := io.Copy(dest, res); err != nil {
		panic(err)
	}

	// // callers should call Finish
	// if bar != nil {
	// 	bar.Finish()
	// }
	return nil
}
