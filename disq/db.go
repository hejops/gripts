package main

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Wrapper over *sqlx.DB
type sql struct {
	// *sqlx.DB // can use db.Close() (etc) directly, but compiler complains

	db *sqlx.DB // the inner DB, with all the typical methods
}

// TODO: fetch pages -> db.db
func dumpDB() {
	// https://github.com/jmoiron/sqlx?tab=readme-ov-file#usage

	// "/users/{USERNAME}/collection/folders/0/releases?per_page=250&page={i}"
	// json -> struct -> row
}

type Row struct {
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

func (s *sql) getArtist(artist string) *[]Row {
	// sql
	// https://pkg.go.dev/database/sql#example-DB.Query-MultipleResultSets

	// sqlx
	// - sql[x] errors if you don't have enough fields to unmarshal your results
	// - struct fields must be uppercase (unlike stdlib sql, which doesn't care)
	// - "direct" unmarshalling into struct allowed (unlike stdlib sql)

	// excluding columns is nonsensical; don't bother
	// https://stackoverflow.com/a/1712243

	rows := []Row{}
	if err := s.db.Select(
		&rows,
		`
		SELECT * FROM collection
		WHERE artist = ?
		`,
		artist,
	); err != nil {
		return nil
	}
	return &rows
}

// TODO: agg on Id -> concat Artist

// // https://www.golinuxcloud.com/golang-enum/
// type AggFunc int
//
// const (
// 	Avg AggFunc = iota
// 	Count
// )
//
// func (f AggFunc) str() string {
// 	return [...]string{"AVG", "COUNT"}[f]
// }

type AvgResult struct {
	Artist    string
	AvgR      float32
	AlbumsStr string
	Albums    []string `db:"-"`
}

// agg on Artist -> avg rating, concat albums
func (s *sql) aggArtistRating() *[]AvgResult {
	// func (s *sql) aggArtistRating() {
	//

	// rows, err := s.db.Queryx("SELECT * FROM collection")
	// if err != nil {
	// 	panic(err)
	// }
	// for rows.Next() {
	// 	results := make(map[string]interface{})
	// 	if err = rows.MapScan(results); err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Println(results)
	// 	// map[artist:Artist ...]
	// 	break
	// }

	// columns must be specified in lowercase
	query := `
		SELECT
			artist,
			AVG(r) avgr,
			GROUP_CONCAT(title, "	") albumsstr
		FROM collection

		GROUP BY artist
		HAVING COUNT(r) > 1
		AND avgr >= 3

		ORDER BY avgr DESC
		-- LIMIT 100
		`

	// r := AvgResult{}
	// artists := []AvgResult{}
	// rows, _ := s.db.Queryx(query)
	// for rows.Next() {
	// 	err := rows.StructScan(&r)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	artists = append(artists, r)
	// }
	// fmt.Println(len(artists))

	// for simple unmarshalling into a []Struct, the above can be written
	// as follows
	rows := []AvgResult{}
	err := s.db.Select(&rows, query)
	if err != nil {
		panic(err)
	}
	// https://github.com/aws-containers/retail-store-sample-app/blob/5b3f349a926fc21145666cad6759ea9725c290d1/src/catalog/repository/mysql_repository.go#L161
	for _, row := range rows {
		row.Albums = strings.Split(row.AlbumsStr, "	")
		row.AlbumsStr = "" // not strictly necessary
	}
	fmt.Println(rows[0])
	return &rows

	// // type AvgResult2 struct {
	// // 	Artist string
	// // 	AvgR   float32
	// // 	Albums []string
	// // }
	// r := AvgResult{}
	// artists := make(map[string]AvgResult)
	// rows, _ := s.db.Queryx(query)
	// // s.db.SelectContext()
	// for rows.Next() {
	// 	err := rows.StructScan(&r)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	artists[r.Artist] = r
	// }
	// fmt.Println(artists["Metallica"])

	// // rows := make(map[string]AvgResult)
	// rows := make(map[string]interface{})
	// if err := s.db.QueryRowx(query).MapScan(rows); err != nil {
	// 	panic(err)
	// }
	// fmt.Println(rows)
	// // map[albums:... artist:... avgr:3]
}
