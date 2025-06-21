package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type YoutubeVideo struct {
	Downloadable
	Url      string
	Title    string
	Uploader string
	Released time.Time
	Age      int
}

func (y YoutubeVideo) url() string { return y.Url }

func (y YoutubeVideo) title() string { return y.Title }

func getYoutubeChannelUploads(uploaderId string) []YoutubeVideo {
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
		if strings.Contains(url, "/shorts/") {
			return true
		}
		t, err := time.Parse(
			time.RFC3339, // not entirely sure
			s.Find("published").First().Text(),
		)
		days := int(time.Since(t).Hours() / 24)
		if days > Cfg.MaxDays {
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

// Given a list of YouTube channels, retrieve URLs of videos posted within the
// last n days
func getYoutubeVideos() []YoutubeVideo {
	// timing seems to vary wildly from 1.2 s to 8 s (58 urls)
	var wg sync.WaitGroup
	videos := []YoutubeVideo{}
	for _, uploaderId := range Cfg.Youtube.Urls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			videos = append(videos, getYoutubeChannelUploads(uploaderId)...)
		}()
	}
	wg.Wait()
	return videos
}
