// A rewrite of a 3+ year-old Bash script. I could never be bothered to
// properly implement 2 loops with different intervals in Bash, but Go makes
// this trivial.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// this should only be used for infallible commands!
func get_cmd_output(cmd string, args ...string) string { // {{{
	bytes, err := exec.Command(cmd, args...).Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(bytes))
} // }}}

func read_file(path string) string { // {{{
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	bytes, err := io.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(bytes))
} // }}}
func get_resp_body(url string) string { // {{{
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(bytes)
} // }}}
func filter(arr []string) []string { // {{{
	// https://josh-weston.scribe.rip/golang-in-place-slice-operations-5607fd90217

	// filter a slice in place without allocating, use two slices with the
	// same backing array
	// see also: https://stackoverflow.com/a/50183212

	i := 0
	for _, v := range arr {
		if v != "" {
			arr[i] = v // overwrite the original slice
			i++
		}
	}
	return arr[:i] // return slice of remaining elements
} // }}}
// count number of newlines in string
func lines(s string) int { // {{{
	l := strings.Count(s, "\n")
	if string(s[len(s)-1]) != "\n" {
		l += 1
	}
	return l
} // }}}

func get_location() string { // {{{
	resp, err := http.Get("https://ipinfo.io")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	var obj map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&obj); err != nil {
		log.Fatal(err)
	}

	// https://go.dev/ref/spec#Type_assertions
	return obj["city"].(string)
} // }}}
func weather(loc string) string { // {{{
	// curl -sL ipinfo.io
	// curl --max-time 1 --fail -sL "wttr.in/$location?format=%C,+%t+(%s)"

	wt := get_resp_body("https://wttr.in/" + loc + "?format=%C,+%t+(%s)")
	return wt
} // }}}

func bat() string {
	path := "/sys/class/power_supply/BAT0/capacity"
	if _, err := os.Stat(path); err != nil {
		// TODO: get laptop battery
		return ""
	}
	return read_file(path)
}

// '+%a %d/%m +%H:%M'
func _time() string {
	// refer to time.Layout
	fmt := "Mon 02/01 15:04" // [Z07]"
	return time.Now().Format(fmt)
}

func sys() string { // {{{
	// free -h | awk 'NR==2 {print $3}' | tr -d i

	mem := get_cmd_output("free", "-h")
	mem = strings.Split(mem, "\n")[1]
	mem = strings.Fields(mem)[2]
	mem = strings.TrimRight(mem, "i")

	// sensors -u | grep temp1_input | sort | tail -n1 | cut -d' ' -f4 | cut -d. -f1

	sensors := get_cmd_output("sensors", "-u")
	var max_temp string
	for _, line := range strings.Split(sensors, "\n") {
		if strings.Contains(line, "temp1_input") {
			temp := strings.Fields(line)[1]
			// don't bother with strconv.Atoi
			if strings.Compare(temp, max_temp) > 0 {
				max_temp = temp
			}
		}
	}
	max_temp, _, _ = strings.Cut(max_temp, ".")
	// fmt.Println(max_temp)

	// top -b -n1 | grep %Cpu | awk '{print $2}'

	var cpu string = "?"
	cpu_out := get_cmd_output("top", "-b", "-n1")
	for _, line := range strings.Split(cpu_out, "\n") {
		if strings.Contains(line, "%Cpu") {
			cpu = strings.Fields(line)[1]
			break
		}
	}

	return fmt.Sprintf("%s%%, %s, %s°C", cpu, mem, max_temp)
} // }}}

func disk() string {
	// df -h / /dev/sda?*
	df := "df -h / /dev/sda?*" // exec.Command does not do shell expansion!
	out := get_cmd_output("sh", "-c", df)
	var arr []string
	for _, line := range strings.Split(out, "\n")[1:] {
		arr = append(arr, strings.Fields(line)[3])
	}
	return strings.Join(arr, " ")
}

func nowplaying() string { // {{{
	status, err := exec.Command("playerctl", "status").Output()
	if err != nil {
		return ""
	}

	np := get_cmd_output(
		"playerctl",
		"metadata",
		"--format",
		"{{ playerName }}: {{ artist }} - {{ title }}",
	)

	if strings.Contains(string(status), "paused") {
		np = "⏸ " + np
	}

	return np
} // }}}

// fetching mail is handled by a systemd timer
func mail() string {
	// NOTMUCH_CONFIG="$HOME/.config/notmuch/config"
	_new := get_cmd_output(
		"notmuch",
		strings.Split("count tag:inbox and tag:unread and date:today", " ")...,
	)
	if _new == "0" {
		return ""
	}
	return _new + " new mail"
}

func updates() string {
	// updates := get_cmd_output("checkupdates")

	bytes, err := exec.Command("checkupdates").Output()
	if err != nil {
		return ""
	}
	return strconv.Itoa(lines(string(bytes))) + " updates"
}

func fast_loop() []string {
	return []string{
		// mail(),
		nowplaying(),
		get_cmd_output("iwgetid", "-r"),
		sys(),
		disk(),
		bat(),
		_time(),
	}
}

// Reserved for making network requests with no action to be taken. Typically,
// this includes weather, stocks, etc.
func slow_loop(loc string) []string {
	return []string{
		weather(loc),
	}
}

type Cacher struct {
	f     func() string
	value string
	count int
}

// Caches are checked slowly when empty, but quickly when non-empty (so that we
// can "clear" the notif)
func (cache *Cacher) update() { // https://gobyexample.com/methods

	// TODO: we should have a time.Duration struct field
	interval := 120 // 10 min / 5 s

	if cache.value != "" {
		fmt.Println(cache.count, cache.value)
		cache.value = cache.f()
	} else if cache.count > interval {
		cache.count = 0
	} else {
		cache.count += 1
	}
}

func main() {
	// init
	name := read_file("/sys/devices/virtual/dmi/id/product_name")
	loc := get_location()
	sep := " | "

	// TODO: investigate slow startup; need some form of logging

	// https://stackoverflow.com/a/40364927
	fast := time.NewTicker(5 * time.Second)
	slow := time.NewTicker(10 * time.Minute)

	a := fast_loop()
	b := slow_loop(loc)

	mail_cache := Cacher{mail, mail(), 0}
	updates_cache := Cacher{updates, updates(), 0}

	for {
		select {
		case <-fast.C:
			a = fast_loop()

			mail_cache.update()
			updates_cache.update()

		case <-slow.C:
			b = slow_loop(loc)
		}

		merged := append(b, a...)
		merged = append(
			[]string{mail_cache.value, updates_cache.value},
			merged...,
		)
		msg := strings.Join(filter(merged), sep)
		msg = name + " > " + msg

		// fmt.Println(msg)

		if err := exec.Command("xsetroot", "-name", msg).Run(); err != nil {
			log.Fatal(err)
		}
	}
}
