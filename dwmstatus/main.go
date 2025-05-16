// Statusbar for `dwm`
//
// A rewrite of a 3+ year-old Bash script. I could never be bothered to
// properly implement 2 loops with different intervals in Bash, but Go makes
// this trivial.
//
// Until I find native Go equivalents, the following executables are required:
//	df, free, iwgetid, top, sensors, playerctl
//
// No non-stdlib imports are allowed.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// Divides sections in the status bar
	Separator = " | "

	// e.g. Mon 26/08 13:44
	TimeFmt = "Mon 02/01 15:04"
	// [Z07]"

	// Only relevant for laptops
	BatteryCapacity = "/sys/class/power_supply/BAT0/capacity"
)

var (
	fastInterval = 5 * time.Second
	slowInterval = 10 * time.Minute

	lastBytes       = netBytes()
	currentLocation = getLocation()

	MachineName = readFile("/sys/devices/virtual/dmi/id/product_name")
	MailCache   = Cacher{
		f:        mail,
		value:    mail(),
		interval: int(slowInterval.Seconds() / fastInterval.Seconds()),
	}
)

type Cacher struct {
	f        func() string
	value    string
	count    int
	interval int

	// interval := 120 // 10 min / 5 s
}

func die(err error) {
	lf, _ := os.OpenFile("/tmp/dwmstatus", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 	panic(err)
	// }
	defer lf.Close()
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

// should be used if any arg contains a space that must be preserved
func getCmdOutput(cmd string, args ...string) string { // {{{
	out, _ := execRawCommand(*exec.Command(cmd, args...))
	return out
} // }}}

// simpler if quoting not required (i.e. when no arg contains a space)
func getCmdOutputLazy(cmd string) string { // {{{
	args := strings.Fields(cmd)
	out, _ := execRawCommand(*exec.Command(args[0], args[1:]...))
	return out
} // }}}

// func getCmdOutputWithFallback(cmd string, fallback string) string { // {{{
// 	args := strings.Fields(cmd)
// 	out, err := execRawCommand(*exec.Command(args[0], args[1:]...))
// 	if err != nil {
// 		return fallback
// 	}
// 	return out
// } // }}}

func readFile(path string) string { // {{{
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
func getRespBody(url string) string { // {{{
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
// Remove empty elements from arr
func filter[T comparable](arr []T) []T { // {{{
	// https://josh-weston.scribe.rip/golang-in-place-slice-operations-5607fd90217

	// filter a slice in place without allocating, use two slices with the
	// same backing array
	// see also: https://stackoverflow.com/a/50183212

	var zero T
	i := 0
	for _, v := range arr {
		if v != zero {
			arr[i] = v // overwrite the original slice
			i++
		}
	}
	return arr[:i] // return slice of remaining elements
} // }}}

type location struct {
	City string
	Lat  float32
	Lon  float32
}

// Determine current location using a geolocation service. Should only be run
// once, on startup.
//
// Note: accuracy can be poor, depending on geolocation, and weather provider
func getLocation() *location { // {{{
	resp, err := http.Get("https://ipinfo.io")
	if err != nil {
		return nil //, errors.New("geolocation failed")
	}
	defer resp.Body.Close()

	// var obj map[string]interface{}
	var obj struct {
		City string
		Loc  string
	}

	err = json.NewDecoder(resp.Body).Decode(&obj)
	if err != nil {
		return nil //, errors.New("weather lookup failed")
	}

	latlon := strings.Split(obj.Loc, ",")
	lat, _ := strconv.ParseFloat(latlon[0], 32)
	lon, _ := strconv.ParseFloat(latlon[1], 32)

	return &location{
		City: obj.City,
		Lat:  float32(lat),
		Lon:  float32(lon),
	} //, nil
} // }}}
// Uses wttr.in for ease of parsing
func weather() string { // {{{
	// curl -sL ipinfo.io
	// curl --max-time 1 --fail -sL "wttr.in/$location?format=%C,+%t+(%s)"

	if currentLocation == nil {
		return ""
	}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"https://api.met.no/weatherapi/locationforecast/2.0/compact?lat=%f&lon=%f",
			currentLocation.Lat,
			currentLocation.Lon,
		),
		nil,
	)
	if err != nil {
		panic(err)
	}
	req.Header.Set("User-Agent", "github.com/hejops/dwmstatus")

	resp, err := http.DefaultClient.Do(req)
	if err != nil { // ignore network errors
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// jq '[.properties | .timeseries[] | .data | .instant | .details | .air_temperature]'

	var x struct {
		Properties struct {
			Timeseries []struct {
				Data struct {
					Instant struct {
						Details struct{ Air_Temperature float32 }
					}
					Next_12_Hours struct {
						Summary struct {
							Symbol_Code string
						}
					}
				}
			}
		}
	}

	if err := json.Unmarshal(body, &x); err != nil {
		panic(err)
	}

	minT := x.Properties.Timeseries[0].Data.Instant.Details.Air_Temperature
	var maxT float32
	for _, t := range x.Properties.Timeseries[:24] {
		minT = min(t.Data.Instant.Details.Air_Temperature, minT)
		maxT = max(t.Data.Instant.Details.Air_Temperature, maxT)
	}

	wt := fmt.Sprintf(
		"%s, %.0f - %.0f°C",
		strings.Split(x.Properties.Timeseries[0].Data.Next_12_Hours.Summary.Symbol_Code, "_")[0],
		minT,
		maxT,
	)

	return wt
} // }}}

func battery() string {
	if _, err := os.Stat(BatteryCapacity); err != nil {
		return ""
	}
	return readFile(BatteryCapacity)
}

// '+%a %d/%m +%H:%M'
func _time() string {
	// refer to time.Layout
	return time.Now().Format(TimeFmt)
}

func sys() string { // {{{

	// SwapUse          0B CachUse        6.9G  MemUse        6.8G MemFree        2.8G
	mem := getCmdOutputLazy("free --line --human --si")
	mem = strings.Fields(mem)[5]
	// TODO: warn if >= 8 G

	// sensors -u | grep temp1_input | sort | tail -n1 | cut -d' ' -f4 | cut -d. -f1

	// parsing the json (sensors -j) is not trivial, due to inconsistent
	// field names
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

	// %Cpu(s):  5.8 us,  1.7 sy,  0.0 ni, 92.5 id,  0.0 wa,  0.0 hi,  0.0 si,  0.0 st

	// 3rd line, but the intent is clearer this way (and performance is
	// essentially identical, despite the 3 extra strings.Contains calls)
	cpu := "?"
	// pid 8 is arbitrary; we just get the summary and ignore the processes
	cpu_out := getCmdOutputLazy("top --batch --iterations=1 --pid=8")
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
	// on some machines, / /dev/sdaX are the same
	df := "df --human-readable --output=avail / /dev/sda?* | uniq"
	out := getCmdOutput("sh", "-c", df)
	var arr []string
	for _, line := range strings.Split(out, "\n")[1:] {
		arr = append(arr, strings.TrimSpace(line))
	}
	return strings.Join(arr, " ")
}

func nowplaying() string { // {{{
	// a marquee is not too hard to implement, but the 5 second interval
	// makes this a moot point
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

// fetching mail is handled by a cronjob
func mail() string {
	cmd := exec.Command(
		"notmuch",
		strings.Fields("count tag:inbox and tag:unread and date:today")...,
	)
	_ = os.WriteFile("/tmp/dwmstatus.log", []byte(strings.Join(os.Environ(), "\n")), os.ModePerm)
	cmd.Env = os.Environ()
	out, err := execRawCommand(*cmd)
	fmt.Println(out)
	switch {
	case err != nil:
		return "mail error"
	case out == "0":
		return ""
	default:
		return out + " new mail"
	}
}

func network() string {
	// name := getCmdOutputWithFallback("iwgetid -r", "No network")

	name, err := execRawCommand(*exec.Command("iwgetid", "-r"))
	if err != nil {
		return "No network"
	}

	nowBytes := netBytes()
	diff := nowBytes - lastBytes
	kbps := diff / int(fastInterval.Seconds()) / 1024

	lastBytes = nowBytes

	return fmt.Sprintf("%s [%d kb/s]", name, kbps)
}

func fastLoop() []string {
	return []string{
		// potentially fallible
		nowplaying(),
		// getCmdOutputWithFallback("iwgetid -r", "No network"),
		network(),
		// the following cmds should all be infallible
		sys(),
		disk(),
		battery(),
		_time(),
	}
}

// Reserved for making network requests with no action to be taken. Typically,
// this includes weather, stocks, etc.
func slowLoop() []string {
	return []string{
		weather(),
	}
}

// Caches are checked slowly (every 10 min) when empty, but quickly (every 5 s)
// when non-empty, so that we can "clear" the notif
func (c *Cacher) update() { // https://gobyexample.com/methods
	if c.value != "" {
		c.value = c.f()
	} else if c.count > c.interval {
		c.value = c.f()
		c.count = 0
	} else {
		c.count += 1
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
		b, e := os.ReadFile(path)
		if e != nil {
			panic(e)
		}

		s := string(b)
		start := strings.IndexRune(s, '(')
		end := strings.IndexRune(s, ')')

		if s[start+1:end] != procName {
			return nil
		}

		pid := strings.Fields(s)[0]
		// TODO: avoid killing our own process
		if p, _ := strconv.Atoi(pid); p == os.Getpid() {
			return nil
		}
		fmt.Println(s, s[start+1:end], pid, os.Getpid())
		if err := exec.Command("kill", pid).Run(); err != nil {
			panic(err)
		}
		fmt.Println("killed", pid, procName)
		// note: we assume at most one instance of dwmstatus is running
		return fs.SkipAll
	})
	if err != nil {
		panic(err)
	}
}

func netBytes() int {
	b, err := exec.Command("ip", "-j", "-s", "link").Output()
	if err != nil {
		panic(err)
	}
	var sources []struct {
		Stats64 struct{ Rx struct{ Bytes int } }
	}
	if err := json.Unmarshal(b, &sources); err != nil {
		panic(err)
	}
	var sum int
	for _, x := range sources {
		sum += x.Stats64.Rx.Bytes
	}
	return sum
}

func main() {
	// checkRestart()

	// https://stackoverflow.com/a/40364927
	fastTick := time.NewTicker(fastInterval)
	slowTick := time.NewTicker(slowInterval)

	a := fastLoop()
	b := slowLoop()

	for {
		select {

		case <-fastTick.C:
			a = fastLoop()
			// note: mail is fetched immediately after login, but
			// update can only be called 10 mins after login.
			// fetching mail should not be the responsibility of
			// this program
			MailCache.update()

		case <-slowTick.C:
			b = slowLoop()
		}

		// [slow] + [fast]
		merged := append(b, a...)
		merged = append([]string{MailCache.value}, merged...)
		msg := strings.Join(filter(merged), Separator)
		msg = MachineName + " > " + msg

		// fmt.Println(msg)

		// TODO: wide chars (e.g. korean) cause date to be truncated

		if err := exec.Command("xsetroot", "-name", msg).Run(); err != nil {
			die(err)
		}
	}
}
