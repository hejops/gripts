// RSS downloader

package main

import (
	"fmt"
	"sync"
	"time"
)

// pb stopped working on go 1.25.1

const (
	MaxConcurrent = 5 // youtube-only
)

func main() {
	LoadConfig()

	var wg sync.WaitGroup
	var numBcOk uint
	wg.Go(func() {
		defer wg.Done()
		for _, label := range getBandcampLabels() {
			for _, r := range label.Releases() {
				if err := r.Download(nil); err == nil {
					numBcOk++
					fmt.Println(numBcOk, "bc:", r.URL)
				}
			}
			time.Sleep(2 * time.Second)
		}
	})

	var wg2 sync.WaitGroup
	var errs sync.Map // concurrent map writes lead to panic
	// durations := make(map[uint][]string)

	for i, v := range getYoutubeVideos() {
		wg2.Go(func() {
			// now := time.Now()

			if err := v.Download(nil); err != nil {
				errs.Store(v.URL, err)
			} else {
				fmt.Println(i, "yt:", v.URL)
			}

			// dur := uint(time.Since(now).Seconds())
			// durations[dur] = append(durations[dur], v.URL)
		})

		// limit concurrent downloads
		if (i+1)%MaxConcurrent == 0 {
			wg2.Wait()
		}
	}

	wg2.Wait()
	wg.Wait()

	for url, e := range errs.Range {
		fmt.Println(url, e)
	}

	// keys := slices.Collect(maps.Keys(durations))
	// slices.Sort(keys)
	// var tot, n uint
	// for _, k := range keys {
	// 	if k > 15 {
	// 		fmt.Println(k, durations[k])
	// 	}
	// 	tot += k
	// 	n++
	// }
	// fmt.Println("average yt download duration:", tot/n)

	// browseFiles(Cfg.Dest)
}
