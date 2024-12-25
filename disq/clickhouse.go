package main

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

func ch_main() {
	conn, err := connect()
	if err != nil {
		panic((err))
	}

	ctx := context.Background()

	schema := `
CREATE TABLE IF NOT EXISTS person (
    first_name String,
    last_name String,
    email String,
    age UInt32
)
ENGINE = ReplacingMergeTree
PRIMARY KEY (email)
`

	if err := conn.Exec(ctx, schema); err != nil {
		panic(err)
	}

	if err := conn.Exec(
		ctx,
		`
		INSERT INTO person (first_name, last_name, email, age) 
		-- clickhouse supports both question mark and $n placeholders
		VALUES (?,?,?,?);
		`,
		"Jason",
		"Moiron",
		"jmoiron@jmoiron.net",
		1,
	); err != nil {
		panic(err)
	}

	// https://clickhouse.com/docs/en/integrations/go#using-structs
	// all tags must be specified, apparently? column name -> struct field
	// inference doesn't seem to work

	var people []struct {
		FirstName string `ch:"first_name"`
		LastName  string `ch:"last_name"`
		Email     string `ch:"email"`
		Age       uint32 `ch:"age"`
	}
	if err := conn.Select(ctx, &people, "SELECT * FROM person FINAL"); err != nil {
		panic(err)
	}
	fmt.Println(people)
}

func connect() (driver.Conn, error) {
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
		return nil, err
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
		return nil, err
	}
	return conn, nil
}
