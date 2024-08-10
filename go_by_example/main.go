// A concise one-file summary of all examples listed on
// https://gobyexample.com/

package main

import (
	"bufio"
	"bytes"
	"cmp"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"maps"
	"math"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"text/template"
	"time"
	"unicode/utf8"
)

const (
	hi  = "hi" // like in Rust, types can usually be inferred
	num = 5
)

// go:embed must be declared at top-level
// paths are always relative to this file; relative paths (e.g. ./) are not
// allowed
//
//go:embed go.mod
var mod string

// types {{{

func types() {
	// assignment
	foo := 1
	_ = foo // shut up, compiler!

	math.Sin(num) // type coercion is performed (no need Rust's `into`)

	// control flow

	for i := 0; i < 3; i++ { // boomer loop: assign, test, increment
		// note: loop variables do not persist beyond the inner scope
		// (like Rust, unlike Python)
	}
	// fmt.Println(i)

	for i := range 3 { // Python-ish loop
		_ = i
	}

	// one statement may precede a condition
	if i := 5; i > 9 || i > 11 {
		fmt.Println(i)
	}

	// Rust/Python `match`
	i := 1
	switch i {
	case 2, 3, 4:
	default:
	}

	// switch with no expression
	switch {
	case 1 > 2:
	}

	var x interface{}
	switch t := x.(type) {
	case bool:
		fmt.Println(t)
	}

	// composite types

	// arrays must be (immutably) sized at compile time, like Rust slices
	_ = [5]int{1, 2, 3, 4, 5}
	_ = [...]int{1, 2, 3, 4, 5} // let compiler infer length
	array := [5]int{1: 1, 4: 3} // if idx given, preceding elements will be 'null'
	fmt.Println("array:", array)

	// slices are dynamically sized, like Rust Vecs
	// "In practice, slices are much more common than arrays."
	// https://go.dev/tour/moretypes/7
	// https://go.dev/blog/slices-intro
	// https://go.dev/blog/slices
	_ = []int{}
	slice := make([]int, 8) // init with allocated capacity
	_ = append(slice, 1)
	_ = len(slice)
	_ = slice[2:5]

	// var c []int
	c := make([]int, len(slice))
	copy(slice, c) // copy requires allocation (var won't work!)
	slices.Equal(slice, c)

	// (hash)maps
	m := make(map[int]rune)
	m[1] = 'a'
	m[2] = 'b'
	val, present := m[5] // poor man's Option<val>
	if present {
		fmt.Println(val)
	}
	m = map[int]rune{1: 'a'}
	maps.Equal(m, m)

	// like Python, any composite type can be `range`d
	// like Lua, keys are always yielded
	for k, x := range "foo" {
		fmt.Println("range-ing a string:", k, x)
		break
	}

	_ = "ABC"[2]
	utf8.RuneCountInString("ABC")
	rune, size := utf8.DecodeLastRuneInString("a")
	fmt.Println(rune, size)

	unsorted := []int{5, 4, 3, 2, 1}
	slices.Sort(unsorted)
	fmt.Println(unsorted)

	letters := []string{"aaa", "bb", "c"}
	// can also be used to sort structs by field
	slices.SortFunc(letters, func(a string, b string) int {
		return cmp.Compare(len(a), len(b))
	})
	fmt.Println(letters)

	// `strings` methods (non-exhaustive)
	s := "foo"
	fmt.Println(strings.Split(s, "o"))
	fmt.Println(strings.ReplaceAll(s, "o", "x"))

	// https://pkg.go.dev/fmt#hdr-Printing
	// `Printf` enforces type checking
	// type mismatches produce warnings, e.g. %!e(int=88)
	m2 := map[int]string{0: "foo", 1: "bar"}
	fmt.Printf("%T\n", m2)  // type only
	fmt.Printf("%v\n", m2)  // pretty
	fmt.Printf("%#v\n", m2) // debug
	// see also: Sprintf (returns string without printing), Fprintf (prints
	// to specified writer)

	_, _ = strconv.ParseUint("10", 0, 32)
	_, _ = strconv.Atoi("10") // shorthand for ParseUint(str, 10, any)

	u, _ := url.Parse("https://gobyexample.com/url-parsing?key=val")
	fmt.Println(u.Scheme)
	fmt.Println(u.Host)
	fmt.Println(u.Path)
	q, _ := url.ParseQuery(u.RawQuery)
	fmt.Println(q)
} // }}}

// funcs {{{

// i don't like implicit typing anyway
func consecutive(a, b, c int) (int, int) {
	_ = c
	return a, b
}
func variadic(args ...int) int { return args[0] }

func funcs() {
	// if you ever need to panic, use `panic` instead of `log.Fatal` for
	// traceback
	// log.Fatal("foo")
	defer func() { // poor man's except
		if e := recover(); e != nil {
			fmt.Println("caught error:", e)
		}
	}()
	panic("foo")

	_, _ = consecutive(1, 2, 3)
	variadic([]int{1, 2, 3}...) // unpack

	closure := func() {
		fmt.Println("closure")
	}
	closure()

	// recursive closure
	var fib func(n int) int // required for fib to be valid within func
	fib = func(n int) int {
		if n < 2 {
			return 1
		}
		return fib(n-1) + fib(n-2)
	}
	fmt.Println("fibo:", fib(7))

	// passing an obj via pointer avoids copying it (improving
	// performance), but may cause the obj to be mutated!
	ptr_mod := func(s *string) {
		*s = "X"
	}
	// ptr_mod(&"A")
	s := "A"
	ptr_mod(&s) // in Rust, let mut s = "A"; ptr_mod(&mut s)
	fmt.Println("modified via ptr:", s, &s)

	// otherwise, objs are always passed as a copy
	str_mod := func(s string) {
		_ = s
		s = "X"
		_ = s
	}
	s = "A"
	str_mod(s)
	fmt.Println("copied:", s, &s)
} // }}}

// structs, interfaces, enums {{{

type Point struct {
	x int
	y int
}

// impl Point { fn new() }
func newPoint(x int, y int) Point {
	return Point{x, y}
}

// https://go.dev/tour/methods/8
//
// There are two reasons to use a pointer receiver.
//
// The first is so that the method can modify the value that its receiver
// points to.
//
// The second is to avoid copying the value on each method call. This can be
// more efficient if the receiver is a large struct, for example.

// pass by value (copy; immutable)
func (p Point) sum() int {
	p.x = 999
	return p.x + p.y
}

// pass by ref, can have side effects!
func (p *Point) sum_ref() int {
	p.x = 999 // -will- modify struct
	return p.x + p.y
}

// no default impls
type Cartesian interface {
	distance() float64
}

// assuming an interface-first design (in practice, i usually do method-first),
// we start from struct S and write in the following order:
//
// Go: interface (type I interface {}) -> method (func (s S) foo() {}) -> interface func (func (i I) bar() {})
// Rust: trait (trait T { fn foo() }) -> method (impl S { fn foo() })
//
// Go warns about unimplemented methods at latest possible stage (when you try
// to call something)
//
// Rust instead warns a bit earlier (before you even call anything)

// interface methods cannot be implemented with pointers!
func (p Point) distance() float64 {
	return math.Sqrt(math.Pow(float64(p.x), 2) + math.Pow(float64(p.y), 2))
}

func distance(c Cartesian) {
	c.distance()
}

// embedding (nested struct)
type Line struct {
	Point // "field without a name"; kind of like `self`?
	other Point
}

// go install golang.org/x/tools/cmd/stringer@latest
//go:generate stringer -type=Enum
// or manually run `stringer -type=Enum`

// enum
type Enum int

const (
	Starting Enum = iota
	InProgress
	Stopped
)

func use_enum(e Enum) {
	switch e {
	case Starting: // do stuff
	case InProgress:
	case Stopped:
	}
}

func structs_enums() {
	fmt.Println(Point{1, 2})
	fmt.Println(newPoint(3, 4))

	p := Point{} // all struct fields have null default values...
	p.x = 5      // and are mutable

	fmt.Println(p.x, p.y) // 5, 0
	_ = p.sum()           // 5, 0
	p.sum_ref()           // 999, 0

	distance(p)

	line := Line{other: Point{1, 1}}
	fmt.Println(
		line.x, // container inherits the fields of the "primary" field
		line.Point.x,
		line.other.x,
	)
	fmt.Println(line.distance())

	// "Embedding structs with methods may be used to bestow interface
	// implementations onto other structs" (whatever that means)
	var f Cartesian = line
	fmt.Println(f.distance())

	use_enum(Starting)
} // }}}

// generics {{{

type element[T any] struct {
	next *element[T]
	val  T
}

// `comparable` is only for ==/!=
func max[T cmp.Ordered](s []T) T {
	var x T
	for _, v := range s {
		if v > x {
			x = v
		}
	}
	return x
}

// to define generic methods, a generic type must first be defined
// https://stackoverflow.com/a/70668559
// having to define a new type is weird, so might as well just use regular functions
type slice[T cmp.Ordered] []T

func (s slice[T]) max() T {
	var x T
	for _, v := range s {
		if v > x {
			x = v
		}
	}
	return x
}

func generics() {
	max([]string{"A", "B", "C"})
	max([]rune{'A', 'B', 'C'})
	max([]int{1, 2, 3})
	slice[int]{1, 2, 3}.max()
} // }}}

// errors{{{

type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

func _errors() {
	if err := errors.New("foo"); err != nil {
		fmt.Println(err)
	}

	nested := fmt.Errorf("foo %w", errors.New("bar"))
	var cust *customError
	if err := func() error { return nested }(); errors.Is(err, nested) {
		fmt.Println(nested)
		// for custom error types (that impl `Error`), use `errors.As`
		// with a pointer/ref
	} else if errors.As(err, &cust) {
		fmt.Println(err)
	}
}

// }}}

// concurrency

func concurrency() {
	// 2 concurrent goroutines (no channel)
	go func() {
		for _, x := range []int{1, 2, 3} {
			fmt.Print(x)
		}
	}()

	go func() {
		for _, x := range []string{"a", "b", "c"} {
			fmt.Print(x)
		}
	}()

	// notice that the output is (usually) not 123abc
	time.Sleep(time.Microsecond * 50)
	fmt.Println()

	c := make(chan string)                // unbuffered channel (blocking)
	go func() { c <- "channel item 1" }() // send to channel

	// // important: attempting to send to an unbuffered channel without a
	// // goroutine = immediate panic
	// // the value "falls" through the channel and has nowhere to go
	// c <- "foo"

	c = make(chan string, 2) // buffered channel (non-blocking)
	c <- "channel item 2"    // now possible to send without goroutine

	go func() { c <- "channel item 3" }()
	// "Channels act as first-in-first-out queues"
	// https://go.dev/ref/spec#Channel_types
	// receive all 3 items from channel
	fmt.Println(<-c) // 2 (sync operations have 'priority')
	fmt.Println(<-c) // 3 (LIFO?)
	fmt.Println(<-c) // 1
	// channel is now empty/waiting (but -not- closed)

	// // warning: trying to receive from an empty (but still open) channel =
	// // deadlock!
	// fmt.Println(<-c)

	// however, closing the channel prevents deadlock
	close(c) // irreversible, should only be called by sender
	// close(c) // closing an already closed channel = panic!

	// `close` only closes the "sending" end:
	//
	// <- C--[1, 2]--C <X- 3
	//
	// receiving from a closed and empty channel = null
	// attempting to send to closed channel = panic

	_, ok := <-c

	fmt.Println("channel open:", ok)

	// https://blog.devtrovert.com/p/go-channels-explained-more-than-just

	// https://substackcdn.com/image/fetch/f_auto,q_auto:good,fl_progressive:steep/https%3A%2F%2Fsubstack-post-media.s3.amazonaws.com%2Fpublic%2Fimages%2F091ce583-f070-46bd-9bfc-1545df5b6c32_1200x675.png

	// open channels will be closed on program exit

	// to ensure a slow goroutine finishes, ensure it sends to a channel
	// with a corresponding receiver; the channel can be one 'dedicated'
	// solely to that goroutine
	done := make(chan bool)
	go func(c chan bool) { // remember: args can be passed to goroutines
		fmt.Println("something slow")
		c <- true
	}(done)
	<-done // terminate the goroutine

	// send- and receive-only channels

	c1 := make(chan string, 1)
	c2 := make(chan string, 1)
	send_only := func(send chan<- string, msg string) {
		fmt.Println("sending:", msg)
		send <- msg
	}
	// directionality is confusing, but only if you think the channel is
	// the one doing the sending/receiving -- it's not. -you- are the one
	// doing it!
	//
	// i.e. when you see `<-chan`, translate that to `me<-chan`
	//
	// https://stackoverflow.com/a/59283546
	receive_and_send := func(recv <-chan string, send chan<- string) {
		fmt.Println("passing...")
		send <- <-recv
	}
	go send_only(c1, "hi")
	go receive_and_send(c1, c2)
	fmt.Println("received:", <-c2)

	// channel + goroutine + select loop
	go func() {
		time.Sleep(time.Millisecond)
		c1 <- "1"
	}()
	go func() {
		time.Sleep(time.Millisecond * 2)
		c2 <- "2"
	}()

	// receive 2 values
	// again, note that trying to receive another value (e.g. i < 3) will
	// cause panic
	for i := 0; i < 2; i++ {
		select {
		case recv1 := <-c1:
			fmt.Println("received (1):", recv1)
		case recv2 := <-c2:
			fmt.Println("received (2):", recv2)
		}
	}

	// enforce a timeout in a select with time.After
	timeout := func() {
		select {
		case recv1 := <-c1:
			fmt.Println("received:", recv1)
		case <-time.After(time.Millisecond):
			fmt.Println("timeout")
		}
	}

	go func() {
		c1 <- "fast"
	}()
	timeout()
	go func() {
		time.Sleep(time.Hour)
		c1 <- "slow"
	}()
	timeout()

	c1 <- "foo"
	s := "foo"
	for i := 0; i < 3; i++ {
		select {
		// note: order of cases is significant; this order means we
		// prefer to receive before sending
		case recv1 := <-c1: // non-blocking receive
			fmt.Println("received from c1:", recv1)
		case c2 <- s: // non-blocking send
			fmt.Println("sent to c2:", s)
		default:
			fmt.Println("no activity")
		}
	}

	// exceeding channel capacity

	d := make(chan string, 2)

	d <- "1"
	fmt.Println("channel full:", len(d) == cap(d)) // https://stackoverflow.com/a/25657232
	d <- "2"
	fmt.Println("channel full:", len(d) == cap(d))

	// // <- C--[1, 2]--C <X- 3
	// d <- "3" // send to full channel = hang (will not iterate at all), but why?
	// // i cannot replicate this hang in main, for some reason (only the
	// // usual deadlock)

	close(d)           // if not closed, range will iterate indefinitely!
	for x := range d { // drain channel
		fmt.Println(x)
	}

	timer := time.NewTimer(time.Millisecond)

	// // stopping a timer before a (non-concurrent) receive event = hang
	// timer.Stop()

	// timers have a receive-only channel; the receive only happens when
	// the time has passed.
	// not much more than a time.Sleep that can be terminated prematurely
	fmt.Println("finished timer", <-timer.C)

	// // again, trying to receive from a closed (?) channel = hang
	// fmt.Println("finished timer", <-timer.C)

	timer = time.NewTimer(time.Millisecond)
	go func() {
		// note: this goroutine will only be evaluated if:
		// 1. given enough time to evaluate (e.g. Sleep)
		// 2. timer is not stopped
		// _ = exec.Command("notify-send", "foo").Run()
		fmt.Println("finished timer", <-timer.C)
	}()
	// time.Sleep(time.Millisecond * 2)
	if timer.Stop() { // violates condition 2 above
		fmt.Println("stopped timer prematurely")
	}

	// tickers have a receive-only channel that constantly supplies time
	// values (at the specified interval) until `Stop`ped
	ticker := time.NewTicker(time.Millisecond)
	go func() {
		for {
			select {
			case t := <-ticker.C:
				fmt.Println("tick", t)
			}
		}
	}()
	time.Sleep(time.Millisecond * 4)
	ticker.Stop()

	// this bit is purely about orchestration; no new syntax/semantics are
	// introduced here

	worker := func(i int, in <-chan int, out chan<- int) {
		for job := range in {
			fmt.Printf("w%d started  %d\n", i, job)
			time.Sleep(time.Millisecond * 50)
			fmt.Printf("w%d finished %d\n", i, job)
			out <- job * 2
		}
	}
	inputs := make(chan int, 5)
	outputs := make(chan int, 5)

	// initialise 3 workers
	for w := 1; w <= 3; w++ {
		go worker(w, inputs, outputs)
	}

	// give the 3 workers 5 jobs
	for i := 1; i <= 5; i++ {
		inputs <- i
	}
	close(inputs) // not necessary but Polite To Do

	// ensure we collect 5 outputs (collect them in a slice or whatever)
	foo := []int{}
	for i := 1; i <= 5; i++ {
		x := <-outputs
		foo = append(foo, x)
	}
	fmt.Println(foo)

	// A WaitGroup waits for a collection of goroutines to finish. The main
	// goroutine calls Add to set the number of goroutines to wait for.
	// Then each of the goroutines runs and calls Done when finished. At
	// the same time, Wait can be used to block until all goroutines have
	// finished.
	var wg sync.WaitGroup

	// tl;dr WaitGroups are like channels, with less boilerplate? remember
	// that with channels must be created, sent to, and received from (and
	// possibly also closed). if you don't need to receive the resulting
	// values, stick with WaitGroups
	//
	// https://stackoverflow.com/a/36057090

	// for a WaitGroup with state n, only n operations/goroutines are
	// allowed to run
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		// could also do wg.Add(5) outside the loop
		go func() {
			// when to use defer: opening a file, marking a
			// WaitGroup as done, etc
			defer wg.Done() // decrement counter when finished
			// if Done is not called, will hang!
			time.Sleep(time.Millisecond * 50)
			fmt.Println("wait group", i)
		}()
	}
	wg.Wait()

	// rate limiting

	reqs := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		reqs <- i
	}
	close(reqs) // required
	// time.Tick is shorthand for time.NewTicker().C; ticker.Stop and
	// ticker.Reset are not available
	limiter := time.Tick(time.Millisecond * 50)
	for req := range reqs {
		<-limiter
		fmt.Println("received request", req, time.Now())
	}

	// rate limiting with burst allowance

	// note: this is a channel of `Time`s. it will contain 3 'quick' times,
	// followed by 'regular' times (at 50 ms intervals)
	burstLimiter := make(chan time.Time, 3)
	for i := 1; i <= 3; i++ {
		burstLimiter <- time.Now()
	}
	// close(burstReqs)
	go func() {
		for t := range time.Tick(time.Millisecond * 50) {
			burstLimiter <- t
		}
	}()
	reqs = make(chan int, 5)
	for i := 1; i <= 5; i++ {
		reqs <- i
	}
	close(reqs) // required
	for req := range reqs {
		<-burstLimiter
		if req <= 3 {
			fmt.Println("received burst request", req, time.Now())
		} else {
			fmt.Println("received request", req, time.Now())
		}
	}

	// atomic structs allow state to be shared across goroutines

	var counter atomic.Uint32
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for x := 0; x < 10; x++ {
				// similar to int++
				counter.Add(1)
			}
		}()
	}
	wg.Wait()
	fmt.Println(counter.Load())

	// to access shared state, use sync.Mutex instead

	type SharedMap struct {
		mutex sync.Mutex
		inner map[string]int
	}
	// Mutex (and structs that contain them) can only be passed by pointer;
	// passing by value is still valid, but will produce compiler warnings
	mapInc := func(sm *SharedMap, k string) {
		sm.mutex.Lock()
		defer sm.mutex.Unlock()
		sm.inner[k]++
	}
	var wg2 sync.WaitGroup
	smap := SharedMap{
		// note: trailing comma required
		inner: map[string]int{"foo": 0, "bar": 0},
	}
	fmt.Println(smap.inner)
	wg2.Add(3)
	go func() { mapInc(&smap, "foo"); wg2.Done() }()
	go func() { mapInc(&smap, "foo"); wg2.Done() }()
	go func() { mapInc(&smap, "bar"); wg2.Done() }()
	wg2.Wait()
	fmt.Println(smap.inner)

	// state sharing can alternatively be achieved with goroutines and
	// channels. however, the example given is woefully confusing, so i
	// skip it
	// https://gobyexample.com/stateful-goroutines
}

func text() {
	// https://pkg.go.dev/text/template
	t := template.Must(template.New("foo").Parse(">>> {{ . }} <<<\n"))
	// Execute will always run until it encounters an error
	// .   arg, as-is
	// .F  field F
	// {{- trim left (only 1x lol)
	if e := t.Execute(os.Stdout, "   foo   "); e != nil {
		fmt.Println(e)
	}

	t = template.Must(template.New("foo").Parse("{{if .}}not {{end}}empty: {{.}}\n"))
	_ = t.Execute(os.Stdout, "")
	_ = t.Execute(os.Stdout, "foo")

	// note: first . is the array, second . is the item
	t = template.Must(template.New("foo").Parse("{{range .}}{{.}} {{end}}\n"))
	_ = t.Execute(os.Stdout, []int{1, 2, 3})

	r := regexp.MustCompile("[abc]")
	fmt.Println(r.MatchString("d"))
	fmt.Println(r.FindAllString("abcde", -1)) // number of matches (-1 = all)
	fmt.Println(string(r.ReplaceAllFunc([]byte("abcde"), bytes.ToUpper)))

	// json is always stored as (minified) strings (which are, in turn,
	// just []byte)
	j, _ := json.Marshal(true)
	fmt.Println(string(j))

	j, _ = json.Marshal([]int{1, 2, 3})
	fmt.Println(string(j))

	j, _ = json.Marshal(struct {
		// all fields must be exported (uppercase)
		Name string
		Age  int `json:"age"` // override json key
	}{
		Name: "John",
		Age:  8,
	})
	fmt.Println(string(j))

	var m map[string]interface{} // `interface` is basically `any`.
	// however, the intended type may still need to be provided (correctly)
	// upon deserialisation.
	jstr := `{"name":"John","age":8.1}`
	_ = json.Unmarshal([]byte(jstr), &m)
	fmt.Println(m["age"])
	fmt.Println(m["age"].(float64))

	// type assertions can be skipped if 'schema' is declared via an empty
	// struct
	// _ = json.Unmarshal([]byte(jstr), &foo{})

	// xml is skipped due to lack of interest; see
	// https://gobyexample.com/xml

	fmt.Println(mod)
}

func time_rng() {
	now := time.Now()
	_ = now.Unix()
	past := time.Date(2001, time.Month(1), 1, 1, 1, 1, 1, time.UTC)
	diff := now.Sub(past) // not absolute
	fmt.Println(past.Add(diff))

	fmt.Println(now.Format(time.RFC3339))
	t, err := time.Parse(time.RFC3339, "2024-08-09T19:56:14+01:00")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(t)

	fmt.Println(rand.IntN(100))
	fmt.Println(rand.IntN(100))
	rand.Float32()
	fmt.Println(rand.New(rand.NewPCG(1, 1)).IntN(100))
	fmt.Println(rand.New(rand.NewPCG(1, 1)).IntN(100))

	s := "hello world"

	hasher := sha256.New()
	hasher.Write([]byte(s))
	fmt.Printf("%x\n", hasher.Sum(nil))
	// (sha) hashes are binary: https://stackoverflow.com/a/59175412
	fmt.Println(hex.EncodeToString(hasher.Sum(nil)))

	enc := base64.StdEncoding.EncodeToString([]byte(s))
	fmt.Println(enc)
	dec, _ := base64.StdEncoding.DecodeString(enc)
	fmt.Println(string(dec))
}

func files() {
	f := "./go.mod"

	b, _ := os.ReadFile(f)
	fmt.Println(strings.Split(string(b), "\n")[0])

	// if you are opening files via Open, it will probably be for
	// performance reasons (i.e. to avoid reading large files)

	fo, _ := os.Open(f)
	defer fo.Close()
	b = make([]byte, 10)
	_, _ = fo.Read(b)
	fmt.Println(string(b)) // typical oob checks apply

	_, _ = fo.Seek(5, io.SeekStart) // jump to 5th byte
	_, _ = fo.Read(b)               // read 10 bytes
	fmt.Println(string(b))

	_, _ = fo.Seek(0, io.SeekCurrent) // maintain cursor
	_, _ = fo.Read(b)                 // read 10 more bytes
	fmt.Println(string(b))

	_, _ = io.ReadAtLeast(fo, b, 10) // another 10 more...
	fmt.Println(string(b))

	// "The bytes stop being valid at the next read call."
	read, _ := bufio.NewReader(fo).Peek(15)
	fmt.Println(string(read))

	tmp := "./foo"
	_ = os.WriteFile(tmp, []byte("hello world"), 0644)
	defer os.Remove(tmp)
	b, _ = os.ReadFile(tmp)
	fmt.Println("created", tmp, "with contents:", string(b))

	// alternatively, Create -> Write[String] -> Sync -> Close, or with
	// NewWriter:
	// _, _ = bufio.NewWriter(fo).WriteString("hello world")

	// // stdin
	// sc := bufio.NewScanner(os.Stdin)
	// for sc.Scan() {
	// 	l := sc.Text() // retrieve the next line
	// 	fmt.Println(">>>", l, "<<<")
	// }
	// // sc.Err()

	// paths
	filepath.Join("a", "b")           // a/b
	filepath.Join("a/b/../c")         // a/c
	filepath.Dir("a/b")               // a
	_, _ = filepath.Rel("a", "a/b/c") // b/c
	// filepath.Ext("a/b.txt")   // txt

	_ = os.MkdirAll("a/b", 0755) // mkdir -p
	defer os.RemoveAll("a")      // rm -rf
	entries, _ := os.ReadDir("a")
	fmt.Println(entries) // [d b/]
	// os.Chdir("a")
	_ = filepath.WalkDir(
		"a",
		// this type signature is required by WalkDir
		func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			fmt.Println(entry, path)
			return nil
		})

	tf, _ := os.CreateTemp("", "file") // /tmp/file12345...
	defer os.Remove(tf.Name())
	td, _ := os.MkdirTemp("", "dir") // /tmp/dir12345...
	defer os.RemoveAll(td)
}

func execution() {
	fmt.Println("args:", os.Args)

	key := flag.String("key", "default", "description")
	fmt.Println("key:", *key)

	fs := flag.NewFlagSet("flagset", flag.ExitOnError)
	fs.Bool("enable", true, "description")

	fmt.Println(os.Getenv("EDITOR"))
	fmt.Println(os.Environ()[0]) // vars can be unpacked with SplitN

	// https://pkg.go.dev/log#pkg-constants
	log.SetFlags(
		log.LstdFlags | // equivalent to Ldate + Ltime
			log.Lmicroseconds |
			log.Lshortfile,
	)
	log.Println("foo")

	// lf, _ := os.Open("log")
	// os.Open is read-only -- https://stackoverflow.com/a/19966217
	lf, _ := os.OpenFile("log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer lf.Close()
	defer os.Remove("log")
	logger := log.New(lf, "logfile", log.LstdFlags)
	logger.Println("foo")

	slog.New(slog.NewJSONHandler(os.Stderr, nil)).Info(
		"foo", // implicit "msg" key
		"key", "val",
	)

	_ = exec.Command("notify-send", "foo").Run()
	ls, _ := exec.Command("ls").Output()
	fmt.Println(string(ls))

	grep := exec.Command("grep", "foo")
	grepIn, _ := grep.StdinPipe() // pipes must be declared before Start
	grepOut, _ := grep.StdoutPipe()
	_ = grep.Start()
	_, _ = grepIn.Write([]byte("foo\nbar"))
	grepIn.Close()
	grepResult, _ := io.ReadAll(grepOut)
	_ = grep.Wait()
	fmt.Println("grep:", string(grepResult))

	// // like the actual OS' `exec`, syscall.Exec completely replaces the
	// // current (go) process
	// lscmd, _ := exec.LookPath("ls")
	// _ = syscall.Exec(lscmd, []string{}, os.Environ())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	done := make(chan bool, 1)
	go func() {
		sig := <-signals
		_ = sig
		fmt.Println("caught signal:", sig)
		done <- true
	}()

	// // uncomment this to block the thread (and see the signal catching in action)
	// <-done
}

func web() {
	resp, _ := http.Get("https://github.com")
	fmt.Println(resp.Status)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(strings.Split(string(body), "\n")[6])

	// // see https://gobyexample.com/http-client
	// bufio.NewScanner(resp.Body)

	http.HandleFunc(
		"/hello",
		func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			defer fmt.Println("context ended")
			select {
			// simulate slow request (and abort)
			case <-time.After(time.Second * 10):
				fmt.Fprintln(w, "hello")
			case <-ctx.Done():
				fmt.Println("error:", ctx.Err())
				http.Error(
					w,
					ctx.Err().Error(),
					http.StatusInternalServerError,
				)
			}
		},
	)
	// http.ListenAndServe(":9999", nil) // server will persist in bg! (pgrep main)
}

func main() {
	types()
	funcs()
	structs_enums()
	generics()
	_errors()
	concurrency()
	text()
	time_rng()
	files()
	execution()
	web()

	// Exit causes all deferred statements to be ignored!
	defer fmt.Println("after exit")
	os.Exit(88)
}
