package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	ApiPrefix = "https://api.discogs.com/"
)

// TODO: this looks interesting, but requires files that are invalid sql https://github.com/Davmuz/gqt

var (
	// The schema requires many double inner joins.
	//go:embed queries/schema.sql
	_schema string
	//go:embed queries/select_random.sql
	_select_random string
	//go:embed queries/select_random_from_artist.sql
	_select_random_from_artist string
)

// Wrapper over *sqlx.DB
type (
	sql struct {
		// *sqlx.DB // can use db.Close() (etc) directly, but compiler complains

		db *sqlx.DB // the inner DB, with all the typical methods
	}

	Artist struct {
		Id   int
		Name string
		// Role string
	}

	Label struct {
		Id   int
		Name string
		// Catno string
	}

	// The result of select_random.sql
	SimpleRow struct {
		Album  string
		Artist string
	}

	// row schema from an old query
	JoinedRow struct {
		Index     string
		Artist    string // Guaranteed to contain only a single artist (though this may change in future)
		Title     string
		Year      int
		Rating    int    `db:"r"` // 1-5
		Genre     string // ", "-delimited
		Label     string
		Id        string // Release id
		DateAdded string `db:"date_added"` // TODO: format 2022-10-11T12:41:04-07:00
		// DateAdded  time.Time `db:"date_added"`

		Img        string
		InstanceId string `db:"iid"`
	}

	AvgResult struct {
		Artist    string
		AvgR      float32
		AlbumsStr string
		Albums    []string `db:"-"` // i forgot the meaning of -
	}
)

// wrapper over sqlx.NamedExec. maybe gorm is easier, but i will hold off for now
func insert(
	tx *sqlx.Tx,
	table string,
	m map[string]any,
) {
	keys := Keys(m)
	var ckeys []string
	for _, k := range keys {
		ckeys = append(ckeys, ":"+k)
	}
	_, err := tx.NamedExec(
		`INSERT OR IGNORE INTO `+table+`
			(`+strings.Join(keys, ",")+`)
			VALUES
			(`+strings.Join(ckeys, ",")+`)
			`,
		m,
	)
	if err != nil {
		panic(err)
	}
}

// excluding columns is just not a thing in sql; don't bother
// https://stackoverflow.com/a/1712243

// method must have no type parameters
// https://stackoverflow.com/a/70668559
// func (s *sql)query[T any]( query string, _ []T) []T {}

func query[T any](
	s *sql,
	_ []T,
	query string,
	args ...any, // not ...string!
) []T { // {{{
	// When making a query, two things are required: the query string, and
	// the expected structure in which to store the result.
	//
	// Annoyingly, there are at least 4 different ways to achieve the same
	// thing in sqlx (none of which are available in database/sql):
	//
	// 1. sqlx.DB.Queryx(SELECT_RANDOM); for rows.Next() { row.StructScan() ... }
	// 2. sqlx.DB.QueryRowx(SELECT_RANDOM).StructScan(&row) // where row is MyRow
	// 3. sqlx.DB.QueryRowx(SELECT_RANDOM).Scan(&col1, ...)
	// 4. sqlx.DB.Select(&rows, SELECT_RANDOM) // where rows is []MyRow
	//
	// Queryx is the "manual" way to perform queries, and is most
	// appropriate when queries are simple, but filtering/transformation
	// potentially complex, and must be performed within Go.
	//
	// QueryRowx removes the need for row iteration, but limits you to a
	// single row. Generally, there is no reason to ever use row.Scan() in
	// sqlx, because passing columns as positional args is brittle.
	//
	// Select has the most ergonomic API, as it completely abstracts away
	// both row iteration (rows.Next()) and unmarshalling
	// (row.StructScan()). As a result, it most closely corresponds a
	// typical json.Unmarshal call.

	var rows []T
	if len(args) > 0 {
		if err := s.db.Select(&rows, query, args...); err != nil {
			panic(err)
		}
	} else {
		if err := s.db.Select(&rows, query); err != nil {
			panic(err)
		}
	}
	return rows
} // }}}

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

// RandomAlbum selects a random album with rating >= 3
func (s *sql) RandomAlbum() []SimpleRow { return query(s, []SimpleRow{}, _select_random) }

func (s *sql) RandomAlbumFromArtist(artist string) []string {
	return query(s, []string{}, _select_random_from_artist, artist)
}
