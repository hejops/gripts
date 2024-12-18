// TUI for Discogs collection

package main

import (
	_ "embed"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// an exercise to work with data via:
// 1. raw SQL (db -> sqlx)
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

// type alias
// type DBWrapper *sqlx.DB

// const DBFile = "./collection.db"
const DBFile = "./collection2.db"

var s sql

//go:embed schema.sql
var schema string

func init() {
	// note: first db connection may be slow to build

	// TODO: Once?
	d, _ := os.Executable() // will be in /tmp for go run
	s.db = sqlx.MustConnect(
		"sqlite3",
		filepath.Join(filepath.Dir(d), DBFile),
	)
}

// TODO: https://fractaledmind.github.io/2023/09/07/enhancing-rails-sqlite-fine-tuning/#pragmas-summary

//go:embed schema.sql
var schema string

func init() {
	s.db.MustExec(schema)
}

func main() {
	s.db.MustExec(schema)
	dumpDB(os.Args[1])
	return

	defer s.db.Close()

	lf, _ := tea.LogToFile("/tmp/didu.log", "didu")
	defer lf.Close()

	m := model{table: sqlToTable(`SELECT * FROM collection`)}

	// all filtering shall be done via sql
	if _, err := tea.NewProgram(&m).Run(); err != nil {
		panic(err)
	}

	// s.aggArtistRating()
	// // TODO: https://github.com/rodaine/table?tab=readme-ov-file#usage

	// rows := *s.getArtist("Johann Sebastian Bach")
	// if rows==nil{}
	// fmt.Println(
	// 	len(rows),
	// 	rows[0],
	// )
}
