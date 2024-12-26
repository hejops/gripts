package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

var ch _clickhouse

type (
	_clickhouse struct{ db driver.Conn }

	ChRow struct {
		AlbumId    uint32    `ch:"album_id"`
		ArtistId   uint32    `ch:"artist_id"`
		ArtistName string    `ch:"artist_name"`
		Title      string    `ch:"title"`
		DateAdded  time.Time `ch:"date_added"`
		Year       uint32    `ch:"year"`
		Rating     byte      `ch:"rating"`
	}
)

func init_clickhouse() { // {{{
	// https://clickhouse.com/docs/en/integrations/go#copy-in-some-sample-code

	var (
		ctx       = context.Background()
		conn, err = clickhouse.Open(&clickhouse.Options{
			// Addr: []string{"localhost:9440"},
			Auth: clickhouse.Auth{
				Database: "default",
				Username: "default",
				// Password: "<DEFAULT_USER_PASSWORD>",
			},
			// ClientInfo: clickhouse.ClientInfo{
			// 	Products: []struct {
			// 		Name    string
			// 		Version string
			// 	}{
			// 		{Name: "an-example-go-client", Version: "0.1"},
			// 	},
			// },
			Debugf: func(format string, v ...interface{}) {
				fmt.Printf(format, v)
			},
			// TLS: &tls.Config{
			// 	InsecureSkipVerify: true,
			// },
		})
	)

	if err != nil {
		panic(err)
	}

	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf(
				"Exception [%d] %s \n%s\n",
				exception.Code,
				exception.Message,
				exception.StackTrace,
			)
		}
		panic(err)
	}

	ch.db = conn

	// Select columns which align with your common filters. If a column is
	// used frequently in WHERE clauses, prioritize including these in your
	// key over those which are used less frequently. Prefer columns which
	// help exclude a large percentage of the total rows when filtered,
	// thus reducing the amount of data which needs to be read.
	//
	// Prefer columns which are likely to be highly correlated with other
	// columns in the table. This will help ensure these values are also
	// stored contiguously, improving compression. GROUP BY and ORDER BY
	// operations for columns in the ordering key can be made more memory
	// efficient.
	//
	// https://clickhouse.com/docs/en/data-modeling/schema-design#choosing-an-ordering-key

	// clickhouse does not support foreign keys at all. it also discourages
	// normalised tables and joins
	//
	// https://clickhouse.com/docs/en/migrations/bigquery#primary-and-foreign-keys-and-primary-index

	// _ = ch.db.Exec(ctx, "DROP TABLE IF EXISTS albums")

	schema := `
CREATE TABLE IF NOT EXISTS albums (
    album_id UInt32 NOT NULL,
    artist_id UInt32 NOT NULL,
    artist_name String NOT NULL,
    title String NOT NULL,
    date_added DateTime NOT NULL,
    year UInt32,
    rating UInt8, -- 0 to 5
)

-- https://clickhouse.com/docs/en/guides/developer/deduplication#using-replacingmergetree-for-upserts
ENGINE = ReplacingMergeTree

PRIMARY KEY (artist_name, rating, album_id)
-- PRIMARY KEY (album_id, rating)
-- PRIMARY KEY (album_id)
`
	// TODO: investigate query speed with different primary key(s)

	if err := ch.db.Exec(ctx, schema); err != nil {
		panic(err)
	}
} // }}}

func ch_test() { // {{{
	// https://clickhouse.com/docs/en/integrations/go#using-structs
	// all tags must be specified, apparently? column name -> struct field
	// inference doesn't seem to work

	var people []struct {
		FirstName string `ch:"first_name"`
		LastName  string `ch:"last_name"`
		Email     string `ch:"email"`
		Age       uint32 `ch:"age"`
	}
	if err := ch.db.Select(
		context.Background(),
		&people,
		"SELECT * FROM person FINAL",
	); err != nil {
		panic(err)
	}
	fmt.Println(people)
} // }}}

func (ch *_clickhouse) InsertAlbum(batch driver.Batch, rel Release) {
	// this API is way nicer than NamedExec, damn
	if err := batch.AppendStruct(&ChRow{
		AlbumId:    uint32(rel.BasicInfo.Id),
		ArtistId:   uint32(rel.BasicInfo.Artists[0].Id),
		ArtistName: rel.BasicInfo.Artists[0].Name,
		Title:      rel.BasicInfo.Title,
		DateAdded:  Must(time.Parse(time.RFC3339, rel.DateAdded)),
		Year:       uint32(rel.BasicInfo.Year),
		Rating:     rel.Rating,
	}); err != nil {
		panic(err)
	}
}
