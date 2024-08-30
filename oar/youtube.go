package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cheggaaa/pb/v3"
	"github.com/wader/goutubedl"
	"golang.org/x/net/context"
)

type YoutubeVideo struct {
	Url      string
	Title    string
	Uploader string
	Released time.Time
	Age      int
}

func getYoutube(uploaderId string) []YoutubeVideo {
	// defer wg.Done()
	url := "https://www.youtube.com/feeds/videos.xml?channel_id=" + uploaderId
	// fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}

	videos := []YoutubeVideo{}

	doc.Find("entry").EachWithBreak(func(i int, s *goquery.Selection) bool {
		url, ok := s.Find("link").Attr("href")
		if !ok {
			return false
		}
		t, err := time.Parse(
			time.RFC3339, // not entirely sure
			s.Find("published").First().Text(),
		)
		days := int(time.Since(t).Hours() / 24)
		if days > 7 {
			return false
		}
		// fmt.Println(days, s.Find("name").First().Text(), url)
		if err != nil {
			panic(err)
		}
		v := YoutubeVideo{
			Url:      url,
			Uploader: s.Find("name").First().Text(),
			Title:    s.Find("title").First().Text(),
			Released: t,
			Age:      int(days),
		}
		videos = append(videos, v)
		return true
	})
	// if len(videos) > 0 {
	// 	fmt.Println(len(videos), uploaderId)
	// }
	return videos
}

func getYoutubeWeek(ids []string) []YoutubeVideo {
	// timing seems to vary wildly from 1.2 s to 8 s (58 urls)
	var wg sync.WaitGroup
	videos := []YoutubeVideo{}
	for _, uploaderId := range ids {
		wg.Add(1)
		go func() {
			defer wg.Done()
			videos = append(videos, getYoutube(uploaderId)...)
		}()
	}
	wg.Wait()
	return videos
}

func (v *YoutubeVideo) download(bar *pb.ProgressBar) {
	fname := v.Title + ".m4a"

	if _, err := os.Stat(fname); err == nil {
		// fmt.Println("already downloaded")
		return
	}

	res, err := goutubedl.Download(
		context.Background(),
		v.Url,
		goutubedl.Options{},
		"bestaudio",
	)
	if err != nil {
		fmt.Printf("failed: %s (%s)\n", v.Url, v.Title)
		return
		// panic(err)
	}
	defer res.Close()

	// _ = os.Remove(fname)
	file, err := os.Create(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// see also (mpb): https://github.com/FantomeBeignet/y2storj/blob/7224cad959e95d9cedcb5f07ec7049764b3ab0ab/y2storj.go#L95
	// https://github.com/vbauerster/mpb#rendering-multiple-bars

	var dest io.Writer
	if bar == nil {
		dest = file
	} else {
		dest = bar.NewProxyWriter(file)
	}

	if _, err := io.Copy(dest, res); err != nil {
		panic(err)
	}
}
