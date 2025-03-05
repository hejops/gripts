package main

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// TODO: this looks interesting, but requires files that are invalid sql https://github.com/Davmuz/gqt

// TODO: https://fractaledmind.github.io/2023/09/07/enhancing-rails-sqlite-fine-tuning/#pragmas-summary

const DBFile = "./collection2.db"

var (
	s sqlite

	// The schema is normalised, and requires many double inner joins.
	//go:embed queries/schema.sql
	_schema string
	//go:embed queries/select_random.sql
	_select_random string
	//go:embed queries/select_random_from_artist.sql
	_select_random_from_artist string
)

// Wrapper over *sqlx.DB
type (
	sqlite struct {
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
		Album  string `ch:"title"`
		Artist string `ch:"artist_name"`
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

func init_sqlite() {
	// note: first db connection may be slow to build. this may not be an
	// issue with clickhouse?

	// try cwd first (go run), then fallback to wherever the binary is
	// ('prod')

	var db_path string
	if _, err := os.Stat(DBFile); err == nil {
		cwd, _ := os.Getwd()
		db_path = filepath.Join(cwd, DBFile)
	} else {
		bin, _ := os.Executable() // binary will be in /tmp for go run
		db_path = filepath.Join(filepath.Dir(bin), DBFile)
	}

	s.db = sqlx.MustConnect("sqlite3", db_path)
	s.db.MustExec(_schema)
}

func (s *sqlite) InsertAlbum(tx *sqlx.Tx, alb Release) {
	s.insert(
		tx,
		"albums",
		map[string]any{
			// go funcs should be used over sql funcs, so that dbs
			// can be more easily swapped out
			"id":         alb.BasicInfo.Id,
			"title":      strings.TrimSpace(alb.BasicInfo.Title),
			"year":       alb.BasicInfo.Year,
			"rating":     alb.Rating,
			"date_added": Must(time.Parse(time.RFC3339, alb.DateAdded)).Unix(),
		},
	)

	for _, a := range alb.BasicInfo.Artists {
		s.insert(
			tx,
			"artists",
			map[string]any{"id": a.Id, "name": a.Name},
		)
		s.insert(
			tx,
			"albums_artists",
			map[string]any{"album_id": alb.BasicInfo.Id, "artist_id": a.Id},
		)
	}
}

// wrapper over sqlx.NamedExec (which guards against sql injection). maybe gorm
// is easier, but i will hold off for now
func (s *sqlite) insert(
	tx *sqlx.Tx,
	table string,
	m map[string]any,
) {
	// https://jmoiron.github.io/sqlx/#namedParams
	// INSERT OR IGNORE INTO albums
	//         (title,year,rating,date_added,id)
	// VALUES
	//         (:title,:year,:rating,:date_added,:id)

	keys := Keys(m)
	var ckeys []string
	for _, k := range keys {
		ckeys = append(ckeys, ":"+k)
	}
	query := `
	INSERT OR IGNORE INTO ` + table + `
		(` + strings.Join(keys, ",") + `)
	VALUES
		(` + strings.Join(ckeys, ",") + `)
	`
	_, err := tx.NamedExec(query, m)
	if err != nil {
		panic(err)
	}
}

// excluding columns is just not a thing in sql; don't bother
// https://stackoverflow.com/a/1712243

// method must have no type parameters
// https://stackoverflow.com/a/70668559
// func (s *sqlite)query[T any]( query string, _ []T) []T {}

func query[T any](
	s *sqlite,
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
	if err := s.db.Select(&rows, query, args...); err != nil {
		panic(err)
	}
	// if len(args) > 0 {
	// 	if err := s.db.Select(&rows, query, args...); err != nil {
	// 		panic(err)
	// 	}
	// } else {
	// 	if err := s.db.Select(&rows, query); err != nil {
	// 		panic(err)
	// 	}
	// }
	return rows
} // }}}

// RandomAlbum selects a random album with rating >= 3
func (s *sqlite) RandomAlbum() []SimpleRow {
	return query[SimpleRow](s, _select_random)
}

func (s *sqlite) RandomAlbumFromArtist(artist string) []string {
	return query[string](s, _select_random_from_artist, artist)
}
