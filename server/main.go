package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
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
	conn.Close()
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		b, _ := exec.Command("playerctl", "status").Output()

		switch strings.TrimSpace(string(b)) {
		case "Playing":
			fmt.Fprintln(w, `<a href="/toggle">Pause</a>`)
		case "Paused":
			fmt.Fprintln(w, `<a href="/toggle">Play</a>`)
		}
	})

	http.HandleFunc("/toggle", func(w http.ResponseWriter, r *http.Request) {
		_ = exec.Command("playerctl", "play-pause").Run()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	log.Printf("starting server on %v:%d\n", LocalIP, PORT)
	err := http.ListenAndServe(fmt.Sprintf(":%d", PORT), nil)
	if err != nil {
		panic(err)
	}
}
