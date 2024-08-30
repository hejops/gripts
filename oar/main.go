package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
)

const MaxConcurrent = 10

func main() {
	// exIo()
	// // Example_multiple()
	// return

	config := LoadConfig()
	// getBandcampWeek(config.Bandcamp.Username)

	videos := getYoutubeWeek(config.Youtube.Urls)
	fmt.Println(len(videos))

	// 2 videos downloaded (+1 ignored) = 350 MB:
	// 49 s (go)
	// 70 s (nogo)
	//
	// 855 MB / 5 = 2m22
	// 855 MB / 10 = 1m57

	// https://github.com/cheggaaa/pb/blob/master/v3/element.go#L36
	// for i/o, counters = total size
	tmpl := `{{string . "prefix" }} {{counters . }} {{speed . "[%s/s]" "[...]" }}`

	bars := []*pb.ProgressBar{}
	for _, v := range videos[:5] {
		bar := pb.Full.New(0)
		bar.Set("prefix", v.Url)
		bar.SetRefreshRate(time.Millisecond * 500)
		bar.SetTemplateString(tmpl)
		// TODO: change style when done
		bars = append(bars, bar)
	}

	pool, err := pb.StartPool(bars...)
	if err != nil {
		panic(err)
	}
	// not supposed to call pool.Start(), apparently

	var wg sync.WaitGroup
	for i, v := range videos[:5] {
		wg.Add(1)

		go func(b *pb.ProgressBar) {
			defer wg.Done()
			b.Start()
			v.download(b)
			b.Finish()
		}(bars[i])

		// limit to 5 concurrent downloads
		if (i+1)%MaxConcurrent == 0 {
			wg.Wait()
		}
	}
	wg.Wait()
	if err := pool.Stop(); err != nil {
		panic(err)
	}
}
