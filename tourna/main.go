// https://en.wikipedia.org/wiki/Single-elimination_tournament

package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"slices"
	"strings"

	"golang.org/x/term"
)

var entries = map[string]int{}

func valuesEq(n int) []string {
	s := []string{}
	for k, v := range entries {
		if v == n {
			s = append(s, k)
		}
	}
	slices.Sort(s)
	return s
}

func valuesGeq(n int) []string {
	s := []string{}
	for k, v := range entries {
		if v >= n {
			s = append(s, k)
		}
	}
	return s
}

func getChar() byte {
	// https://stackoverflow.com/a/70627571

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	b := make([]byte, 1)
	_, err = os.Stdin.Read(b)
	if err != nil {
		panic(err)
	}

	return b[0]
}

func main() {
	b, err := os.ReadFile("./list.txt")
	if err != nil {
		panic(err)
	}
	for line := range strings.SplitSeq(string(b), "\n") {
		// TODO: round down to nearest exponent of 2; all lines must be collected first
		if line == "" {
			break
		}
		x := strings.Split(line, "|")
		entries[x[0]] = 0
	}

	x := int(math.Log2(float64(len(entries)))) // 2^x = len

	for floor := range x {
		candidates := valuesGeq(floor)

		// https://stackoverflow.com/a/46185753
		rand.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})

		for i := 0; i < len(candidates); i += 2 {
			fmt.Println(candidates[i])
			fmt.Println(candidates[i+1])
			fmt.Println()
			x := i
			if getChar() == 'j' {
				x++
			}
			entries[candidates[x]]++
		}

	}

	for n := x; n > 0; n-- {
		fmt.Println(n, valuesEq(n))
	}
}
