package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	ApiPrefix = "https://api.discogs.com/"
)

type (
	Release struct {
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
		Releases   []Release
	}

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
		}
		resp.Body.Close()

		if pg > x.Pagination.Pages {
			break
		}

		// note: these timings are not 100% fair since they include http get
		// and Println
		// sql trim 0.7 - 1.1 s
		// go trim 0.7 - 1.1 s

		tx := s.db.MustBegin()
		batch, err := ch.db.PrepareBatch(context.Background(), "INSERT INTO albums")
		if err != nil {
			panic(err)
		}

		for _, rel := range x.Releases {
			s.InsertAlbum(tx, rel)
			ch.InsertAlbum(batch, rel)
		}

		if err := tx.Commit(); err != nil {
			panic(err)
		}
		if err := batch.Send(); err != nil {
			panic(err)
		}

		fmt.Println(pg, "ok")
		time.Sleep(time.Second)

	}
} // }}}
