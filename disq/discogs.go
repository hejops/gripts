package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const ApiPrefix = "https://api.discogs.com/"

type Release struct {
	DateAdded  string `json:"date_added"` // 2022-10-23T15:45:21-07:00
	InstanceId int    `json:"instance_id"`
	Rating     int
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

// Write to sqlite. Authorization is required.
func dumpDB(user string) { // {{{
	u, _ := url.Parse(ApiPrefix)
	path := fmt.Sprintf("/users/%s/collection/folders/0/releases", user)
	u = u.JoinPath(path) // note: url.JoinPath can error, but URL.JoinPath does not

	v := url.Values{}
	v.Set("per_page", "250")

	prog, _ := os.Executable()
	b, err := os.ReadFile(filepath.Dir(prog) + "/.env")
	if err != nil {
		panic(err)
	}

	var token string
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "DISCOGS_TOKEN") {
			token = strings.Split(line, "=")[1]
			break
		}
	}
	if token == "" {
		panic("DISCOGS_TOKEN not found")
	}

	maxPg := math.MaxUint16
	ids := make(map[int]int)

	for pg := 1; pg <= maxPg; pg++ {
		v.Set("page", strconv.Itoa(pg))
		u.RawQuery = v.Encode()

		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			panic(err)
		}

		req.Header.Add("Authorization", "Discogs token="+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}

		var x struct {
			Pagination struct{ Pages int }
			Releases   []Release
		}
		if err := json.Unmarshal(Must(io.ReadAll(resp.Body)), &x); err != nil {
			panic(err)
		}
		_ = resp.Body.Close()

		// on hitting rate limit, discogs returns a valid null-ish
		// resp, which is actually an error
		if x.Pagination.Pages == 0 {
			panic("hit rate limit")
			// pg--
		}
		maxPg = x.Pagination.Pages

		sqliteTx := s.db.MustBegin()
		// chTx, err := ch.db.PrepareBatch(context.Background(), "INSERT INTO albums")
		// if err != nil {
		// 	panic(err)
		// }

		// fmt.Println(pg)
		for _, rel := range x.Releases {
			s.InsertAlbum(sqliteTx, rel)
			// ch.InsertAlbum(chTx, rel)

			ids[rel.BasicInfo.Id]++

			// panic(rel)
			if rel.Rating < 1 || rel.Rating > 5 {
				panic("got 0 rating; no discogs token supplied?")
			}

		}

		if err := sqliteTx.Commit(); err != nil {
			panic(err)
		}
		// if err := chTx.Send(); err != nil {
		// 	panic(err)
		// }

		// fmt.Printf("%d/%d ok\n", pg, x.Pagination.Pages)
		time.Sleep(2 * time.Second)

	}

	// if slices.Max(slices.Collect(maps.Values(ids))) > 1 {
	for k, v := range ids {
		if v > 1 {
			fmt.Println(k)
		}
	}
} // }}}
