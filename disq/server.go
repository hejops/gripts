package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

const PORT = 3838

var LocalIP string

func init() {
	// https://stackoverflow.com/a/37382208
	conn, err := net.Dial("udp", "1:") // arbitrary non-zero addr
	if err != nil {
		log.Fatal(err)
	}
	LocalIP = strings.Split(conn.LocalAddr().String(), ":")[0]
	fmt.Println(LocalIP)
	conn.Close()
}

func listen() {
	// http.Handle("/", templ.Handler(tableRows(foo())))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// HandleFunc + Component.Render ensures content is dynamically rendered
		err := tableRows(ch.RandomAlbum(10)).Render(r.Context(), w)
		if err != nil {
			fmt.Fprintln(w, err)
		}
	})

	log.Printf("starting server on http://%v:%d\n", LocalIP, PORT)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", PORT), nil); err != nil {
		panic(err)
	}
}
