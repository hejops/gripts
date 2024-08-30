package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const BC_DATE_FMT = "January 2, 2006"

type BandcampLabel struct {
	ArtId    int `json:"art_id"`
	BandId   int `json:"band_id"`
	Location string
	Name     string

	UrlHints struct {
		Subdomain string
	} `json:"url_hints"`
	// so that's how you do it...
}

type BandcampRelease struct {
	Artist   string
	Title    string
	Released time.Time // may be in the future
	Age      int       // days
}

func getBandcampLabels(username string) []BandcampLabel { // {{{

	url := "https://bandcamp.com/" + username
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}

	// fmt.Println(doc.First().Html())

	// <button id="follow-unfollow_12345" type="button" class="follow-unfollow ">
	var id string
	doc.Find("button").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if v, ex := s.Attr("id"); ex {
			v = strings.Split(v, "_")[1]
			id = v
			return false
		}
		return true
	})
	// fmt.Println(id)

	t := strconv.FormatInt(time.Now().Unix(), 10) // Itoa does not accept int64?

	b, err := json.Marshal(map[string]string{
		"fan_id": id,
		// hack: setting older_than_token to current epoch, and count
		// to some large int returns all items in a single resp
		// "older_than_token": "9999999999:9999999999",
		"older_than_token": t + ":" + t,
		"count":            "9999",
	})
	if err != nil {
		panic(err)
	}

	postResp, err := http.Post(
		"https://bandcamp.com/api/fancollection/1/following_bands",
		"application/json",
		bytes.NewBuffer(b),
	)
	if err != nil {
		panic(err)
	}
	defer postResp.Body.Close()

	bb, err := io.ReadAll(postResp.Body)
	if err != nil {
		panic(err)
	}
	var x struct {
		Followeers []BandcampLabel // 'followeers' is not a typo
	}
	_ = json.Unmarshal(bb, &x)
	return x.Followeers
} // }}}

func (l *BandcampLabel) getReleases(maxAge int) []BandcampRelease { // {{{
	url := fmt.Sprintf("https://%s.bandcamp.com", l.UrlHints.Subdomain)
	resp := getRetry(url + "/music")
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}

	h, _ := doc.First().Html()
	if !strings.Contains(h, "/album/") {
		// e.g. https://horribleroom.bandcamp.com/
		return []BandcampRelease{}
	}

	albumUrls := []string{}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		urlpath, ex := s.Attr("href")
		if ex && strings.Contains(urlpath, "/album/") {
			switch urlpath[0] == '/' {
			case true: // relative
				albumUrls = append(albumUrls, url+urlpath)
			case false:
				urlpath, _, _ := strings.Cut(urlpath, "?")
				albumUrls = append(albumUrls, urlpath)
			}
		}
	})

	releases := []BandcampRelease{}
	for _, albumUrl := range albumUrls {
		r := extractRelease(albumUrl)
		if r.Age < 0 {
			panic("not impl")
		}
		if r.Age > maxAge {
			break
		}
		releases = append(releases, r)
	}

	return releases
} // }}}

func extractRelease(url string) BandcampRelease { // {{{
	resp := getRetry(url)
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}

	// h, _ := doc.First().Html()
	// if !strings.Contains(h, "tralbumData") {
	// 	panic("not an album? " + url)
	// }

	title, _ := doc.Find("title").First().Html()
	fields := strings.Split(title, " | ")

	// <div class="tralbumData tralbum-credits">
	//
	//     released July 26, 2024

	var textDate string
	doc.Find("div").EachWithBreak(func(i int, s *goquery.Selection) bool {
		v, ex := s.Attr("class")
		if ex && v == "tralbumData tralbum-credits" {
			textDate = s.Text()
			return false
		}
		return true
	})
	textDate = strings.TrimSpace(textDate)
	textDate = textDate[strings.Index(textDate, " ")+1:]
	t, _ := time.Parse(BC_DATE_FMT, textDate)
	days := time.Since(t).Hours() / 24

	return BandcampRelease{
		Title:    fields[0],
		Artist:   fields[1],
		Released: t,
		Age:      int(days),
	}
} // }}}

func getBandcampWeek(username string) []BandcampRelease {
	labels := getBandcampLabels(username)
	var wg sync.WaitGroup
	releases := []BandcampRelease{}
	for i, label := range labels {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := label.getReleases(7)
			if len(r) == 0 {
				return
			}
			fmt.Println(i, label.Name, len(r))
			releases = append(releases, r...)
		}()
	}
	wg.Wait()
	fmt.Printf("found %d releases from %d labels\n", len(releases), len(labels))
	return releases
}
