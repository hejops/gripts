package main

import (
	"fmt"
	"net/http"
	"time"
)

func readline(prompt string) string {
	fmt.Print(prompt + ": ")
	var input string
	_, _ = fmt.Scanln(&input)
	if input == "" {
		panic("input cannot be empty")
	}
	return input
}

// Warning: invalid urls (e.g. empty strings) will probably fail silently
func getRetry(_url string) *http.Response {
	// interval <> total time for 257 labels
	// 60s -- 8m19
	// 30s -- 7m24
	// 15s -- 25m33 (panic)

	// 60s
	// 261/261 8m14s
	// 268/268 12m41s

	i := 0
	for {
		resp, err := http.Get(_url)
		switch {
		case i >= 100:
			panic(err)
		case err == nil && resp.StatusCode == 200:
			return resp
		// case i > 0 && i%10 == 0:
		// 	fmt.Println("tried", i, url)
		// 	fallthrough
		default:
			time.Sleep(time.Minute)
			// time.Sleep(time.Second * 30)
			i++
		}
	}
}

// // Pretty-print arbitrary http (json) response without needing to know its
// // schema
// //
// // Warning: resp will be closed
// func debugResponse(resp *http.Response) {
// 	defer resp.Body.Close()
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		panic(err)
// 	}
// 	var data map[string]any
// 	if err := json.Unmarshal(body, &data); err != nil {
// 		panic(err)
// 	}
// 	x, _ := json.MarshalIndent(data, "", "    ")
// 	fmt.Println(string(x))
// }
