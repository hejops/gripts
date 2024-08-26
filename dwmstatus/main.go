// A rewrite of a 3+ year-old Bash script. I could never be bothered to
// properly implement 2 loops with different intervals in Bash, but Go makes
// this trivial.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const Separator = " | "

var (
	MachineName = read_file("/sys/devices/virtual/dmi/id/product_name")
	Location    = getLocation()
	MailCache   = Cacher{f: mail, value: mail()}
)

func die(err error) {
	lf, _ := os.OpenFile("/tmp/dwmstatus", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 	panic(err)
	// }
	log.New(lf, "dwmstatus", log.LstdFlags).Println(err)
}

// note: error checking is not really done in the Cmd-related functions

// internal; should only be used if env is required. otherwise, use
// getCmdOutput
func execRawCommand(cmd exec.Cmd) (string, error) { // {{{
	bytes, err := cmd.Output()
	if err != nil {
		return "", err
		// die(err)
	}
	return strings.TrimSpace(string(bytes)), nil
} // }}}

// if any arg contains a space
func getCmdOutput(cmd string, args ...string) string { // {{{
	out, _ := execRawCommand(*exec.Command(cmd, args...))
	return out
} // }}}

// simpler if quoting not required (i.e. when no arg contains a space)
func getCmdOutputLazy(cmd string) string { // {{{
	args := strings.Split(cmd, " ")
	out, _ := execRawCommand(*exec.Command(args[0], args[1:]...))
	return out
} // }}}

func getCmdOutputWithFallback(cmd string, fallback string) string { // {{{
	args := strings.Split(cmd, " ")
	out, err := execRawCommand(*exec.Command(args[0], args[1:]...))
	if err != nil {
		return fallback
	}
	return out
} // }}}

func read_file(path string) string { // {{{
	f, err := os.Open(path)
	if err != nil {
		die(err)
	}
	defer f.Close()
	bytes, err := io.ReadAll(f)
	if err != nil {
		die(err)
	}
	return strings.TrimSpace(string(bytes))
} // }}}
// returns empty string on Get failure
func get_resp_body(url string) string { // {{{
	// resp, err := http.Get(url)
	c := http.Client{Timeout: time.Second * 3}
	resp, err := c.Get(url)
	if err != nil {
		log.Println("Get failed:", url)
		return ""
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		die(err)
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

// // count number of newlines in string
// func lines(s string) int { // {{{
// 	l := strings.Count(s, "\n")
// 	if string(s[len(s)-1]) != "\n" {
// 		l += 1
// 	}
// 	return l
// } // }}}

// note: accuracy tends to be poor
// returns empty string on encountering any error
func getLocation() string { // {{{
	resp, err := http.Get("https://ipinfo.io")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	var obj map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&obj); err != nil {
		return ""
	}

	// https://go.dev/ref/spec#Type_assertions
	return obj["city"].(string)
} // }}}
func weather(loc string) string { // {{{
	// curl -sL ipinfo.io
	// curl --max-time 1 --fail -sL "wttr.in/$location?format=%C,+%t+(%s)"

	if loc == "" {
		return ""
	}

	wt := get_resp_body("https://wttr.in/" + url.QueryEscape(loc) + "?format=%C,+%t+(%s)")
	if wt == "" {
		log.Println("failed to get weather for", loc)
	}
	if strings.Contains(wt, "Sorry") {
		return ""
	}
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

	mem := getCmdOutputLazy("free -Lh")
	mem = strings.Fields(mem)[5]
	mem = strings.TrimRight(mem, "i")

	// sensors -u | grep temp1_input | sort | tail -n1 | cut -d' ' -f4 | cut -d. -f1

	sensors := getCmdOutputLazy("sensors -u")
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
	cpu_out := getCmdOutputLazy("top -b -n1")
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
	// exec.Command does not do shell expansion!
	df := "df --human-readable --output=avail / /dev/sda?*"
	out := getCmdOutput("sh", "-c", df)
	var arr []string
	for _, line := range strings.Split(out, "\n")[1:] {
		arr = append(arr, strings.TrimSpace(line))
	}
	return strings.Join(arr, " ")
}

func nowplaying() string { // {{{
	status, err := exec.Command("playerctl", "status").Output()
	if err != nil {
		return ""
	}

	np := getCmdOutput(
		"playerctl",
		"metadata",
		"--format",
		"{{ playerName }}: {{ artist }} - {{ title }}",
	)

	if strings.Contains(string(status), "Paused") {
		np = "⏸ " + np
	}

	return np
} // }}}

// fetching mail is handled by a systemd timer
func mail() string {
	cmd := exec.Command(
		"notmuch",
		strings.Split("count tag:inbox and tag:unread and date:today", " ")...,
	)
	cmd.Env = os.Environ()
	out, _ := execRawCommand(*cmd)
	if out == "0" {
		return ""
	}
	return out + " new mail"
}

// func updates() string {
// 	bytes, err := exec.Command("checkupdates").Output() // exit code 2 if no updates
// 	if err != nil {
// 		return ""
// 	}
// 	return strconv.Itoa(lines(string(bytes))) + " updates"
// }

func fast_loop() []string {
	return []string{
		// potentially fallible
		nowplaying(),
		getCmdOutputWithFallback("iwgetid -r", "No network"),
		// the following cmds should all be infallible
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
		cache.value = cache.f()
	} else if cache.count > interval {
		cache.value = cache.f()
		cache.count = 0
	} else {
		cache.count += 1
	}
	// fmt.Println(cache.count, cache.value)
}

// Check if an existing dwmstatus instance is running, and, if found, kill it
// before starting the new instance.
func checkRestart() {
	// https://github.com/mitchellh/go-ps/blob/master/process_linux.go

	procName := filepath.Base(os.Args[0]) // either "main" (go run) or "dwmstatus" (binary)
	if procName == "main" {
		fmt.Println("exiting")
		os.Exit(0)
	}

	err := filepath.WalkDir("/proc", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() || filepath.Base(path) != "stat" {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}

		s := string(b)
		start := strings.IndexRune(s, '(')
		end := strings.IndexRune(s, ')')

		if s[start+1:end] != procName {
			return nil
		}

		pid := strings.Fields(s)[0]
		if err := exec.Command("kill", pid).Run(); err != nil {
			panic(err)
		}
		fmt.Println("killed", pid, procName)
		return fs.SkipAll
	})
	if err != nil {
		panic(err)
	}
}

func main() {
	checkRestart()

	// https://stackoverflow.com/a/40364927
	fast := time.NewTicker(5 * time.Second)
	slow := time.NewTicker(10 * time.Minute)

	a := fast_loop()
	b := slow_loop(Location)

	for {
		select {
		case <-fast.C:
			a = fast_loop()

			MailCache.update()
			// updates_cache.update()

		case <-slow.C:
			b = slow_loop(Location)
		}

		// [slow] + [fast]
		merged := append(b, a...)
		merged = append(
			[]string{
				MailCache.value,
				// updates_cache.value,
			},
			merged...,
		)
		msg := strings.Join(filter(merged), Separator)
		msg = MachineName + " > " + msg

		// fmt.Println(msg)

		// TODO: wide chars (e.g. korean) cause date to be truncated

		if err := exec.Command("xsetroot", "-name", msg).Run(); err != nil {
			die(err)
		}
	}
}
