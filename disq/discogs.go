package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	ApiPrefix = "https://api.discogs.com/"
)

func dumpDB(user string) { // {{{
	// https://github.com/jmoiron/sqlx?tab=readme-ov-file#usage

	u, _ := url.Parse(ApiPrefix)

	path := fmt.Sprintf("/users/%s/collection/folders/0/releases", user)
	u = u.JoinPath(path) // note: url.JoinPath can error, but URL.JoinPath does not

	v := url.Values{}
	v.Set("per_page", "250")

	var x struct {
		Pagination struct{ Pages int }
		Releases   []struct {
			DateAdded  string `json:"date_added"` // 2022-10-23T15:45:21-07:00
			InstanceId int    `json:"instance_id"`
			Rating     byte
			BasicInfo  struct {
				Id       int // resource_url is derived from this
				MasterId int `json:"master_id"` // may be 0; master_url is derived from this
				Title    string
				Year     int
				Genres   []string
				Styles   []string

				Artists []Artist
				Labels  []Label
				// Formats []Format
			} `json:"basic_information"`
		}
	}

	tx := s.db.MustBegin()

	for pg := 1; ; pg++ {
		v.Set("page", strconv.Itoa(pg))
		u.RawQuery = v.Encode()

		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			panic(err)
		}

		// no auth required, apparently

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
		// defer resp.Body.Close()

		err = json.Unmarshal(Must(io.ReadAll(resp.Body)), &x)
		if err != nil {
			panic(err)
			// break
		}
		resp.Body.Close()

		// note: these timings are not 100% fair since they include http get
		// and Println
		// sql trim 0.7 - 1.1 s
		// go trim 0.7 - 1.1 s

		for _, alb := range x.Releases {
			insert(
				tx,
				"albums",
				map[string]any{
					// go funcs should be used over sql
					// funcs, so that dbs can be more
					// easily swapped out
					"id":         alb.BasicInfo.Id,
					"title":      strings.TrimSpace(alb.BasicInfo.Title),
					"year":       alb.BasicInfo.Year,
					"rating":     alb.Rating,
					"date_added": Must(time.Parse(time.RFC3339, alb.DateAdded)).Unix(),
				},
			)

			for _, a := range alb.BasicInfo.Artists {
				insert(
					tx,
					"artists",
					map[string]any{"id": a.Id, "name": a.Name},
				)
				insert(
					tx,
					"albums_artists",
					map[string]any{"album_id": alb.BasicInfo.Id, "artist_id": a.Id},
				)
			}

		}

		if pg == x.Pagination.Pages {
			break
		}

		pg++
		time.Sleep(time.Second)

	}

	err := tx.Commit()
	if err != nil {
		panic(err)
	}
} // }}}
