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
	"github.com/cheggaaa/pb/v3"
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
	Downloadable
	Artist   string
	Title    string
	Url      string
	Released time.Time // may be in the future
	Age      int       // days
	Label    string
}

func getBandcampLabels() []BandcampLabel { // {{{

	url := "https://bandcamp.com/" + Cfg.Bandcamp.Username
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
		// HACK: setting older_than_token to current epoch, and count
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

func (l *BandcampLabel) getReleases() []BandcampRelease { // {{{
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
		// TODO: if albumUrl in db, can skip
		r := newBandcampRelease(albumUrl)
		// r := BandcampRelease{}.fromUrl(albumUrl) // seems un-idiomatic
		if r.Age < 0 { // future
			continue
		}
		if r.Age > Cfg.MaxDays {
			break
		}
		releases = append(releases, r)
	}

	return releases
} // }}}

// func (r BandcampRelease) fromUrl(url string) BandcampRelease {
// 	return BandcampRelease{}
// }

// Parse the raw contents of bandcamp album HTML
func newBandcampRelease(url string) BandcampRelease { // {{{
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
	// fmt.Println(h)

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

	// TODO: write to db

	rel := BandcampRelease{
		Title:    fields[0],
		Url:      url,
		Released: t,
		Age:      int(days),
	}

	if len(fields) == 2 {
		// v/a releases don't have artist
		rel.Label = fields[1]
	} else {
		rel.Artist = fields[1]
		rel.Label = fields[2]
	}

	return rel
} // }}}

// Retrieve all bandcamp releases from the last 7 days. Slow due to
// rate-limiting.
func getBandcampReleases() []BandcampRelease {
	labels := getBandcampLabels()
	var wg sync.WaitGroup
	releases := []BandcampRelease{}
	// labelMap := make(map[BandcampLabel][]BandcampRelease)
	// bar := pb.Full.Start(len(labels))

	// single bar

	// counters = number of processed items
	barTemplate := `{{string . "prefix"}} {{counters .}}/{{string . "total"}} {{etime .}}` // {{speed . "[%s/s]" ""}}`

	bar := pb.Full.New(0).
		Set("prefix", "bandcamp").
		Set("total", len(labels)).
		SetRefreshRate(time.Second).
		SetTemplateString(barTemplate)
	bar.Start()
	for _, label := range labels {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := label.getReleases()
			bar.Increment()
			if len(r) == 0 {
				return
			}
			releases = append(releases, r...)
			// labelMap[label] = releases
		}()
	}
	wg.Wait()
	bar.Finish()
	return releases
}
