// Browse a Discogs collection.
//
// The MVP dumps a user's collection and makes a query to select a random
// album. This fits my primary use case of "I'm away from my local media
// storage and want to listen to something I previously liked". This has
// already been implemented, albeit in a hacky mix of go, jq and bash.
//
// A more sophisticated program would enable interactive browsing and filtering
// of the collection, abstracting away complex logic with sql queries.

package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// an exercise to work with data via:
// 1. raw SQL (db -> sqlx/clickhouse)
// 2. SQL-ish (csv? -> go-duckdb) -- unusable!
// 3. PRQL (csv -> duckdb-prql) -- this would just rely on go-duckdb, so no
// 4. ORM (db -> gorm)

// 1. https://github.com/jmoiron/sqlx?tab=readme-ov-file#usage
// 1. https://jmoiron.github.io/sqlx/#connecting
// 2. https://duckdb.org/docs/guides/file_formats/csv_import.html
// 2. https://duckdb.org/docs/sql/dialect/friendly_sql.html (optional)
// 2. https://github.com/marcboeker/go-duckdb/blob/main/examples/simple/main.go
// 3. https://github.com/ywelsch/duckdb-prql?tab=readme-ov-file#running-the-extension
// 4. https://gorm.io/docs/connecting_to_the_database.html#SQLite

// TODO: integrate with youtube, mpv

var s sql

func init() {
	// note: first db connection may be slow to build. this may not be an
	// issue with clickhouse?

	// try cwd first (go run), then fallback to wherever the binary is
	// ('prod')

	const DBFile = "./collection2.db"

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

// TODO: https://fractaledmind.github.io/2023/09/07/enhancing-rails-sqlite-fine-tuning/#pragmas-summary

func main() {
	// dumpDB(os.Args[1])
	// return

	defer s.db.Close()

	fmt.Println(s.RandomAlbum())
	fmt.Println(s.RandomAlbumFromArtist("Metallica"))

	// m := model{table: sqlToTable(`SELECT * FROM collection`)}
	// // all filtering shall be done via sql
	// if _, err := tea.NewProgram(&m).Run(); err != nil {
	// 	panic(err)
	// }

	// s.aggArtistRating()
	// // TODO: https://github.com/rodaine/table?tab=readme-ov-file#usage

	// rows := *s.getArtist("Johann Sebastian Bach")
	// if rows==nil{}
	// fmt.Println(
	// 	len(rows),
	// 	rows[0],
	// )
}
