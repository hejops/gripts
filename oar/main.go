// RSS downloader

package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
)

const (
	MaxConcurrent = 5 // youtube-only

	Spinner = `{{ cycle . "⠋" "⠙" "⠹" "⠸" "⠼" "⠴" "⠦" "⠧" "⠇" "⠏" }}`
	// https://github.com/hetznercloud/cli/blob/b59cfbfdb338bad3a7b80c0569248a8e3abaad01/internal/ui/progress_terminal.go#L34

	BarTemplate = `{{string . "prefix"}} {{counters .}} {{bar .}} {{rtime .}} {{speed . "[%s/s]" ""}}`
	// https://github.com/cheggaaa/pb/blob/master/v3/element.go#L36
	// for i/o, counters = total size

	SpinnerTemplate = `{{string . "prefix"}} ` + Spinner
)

func main() {
	LoadConfig()

	for i, r := range getBandcampReleases() {

		// note: concurrent downloads are avoided because rate-limiting
		// would be almost guaranteed

		_ = r.Download(pb.Full.New(0))
		fmt.Println("ok:", i, r.Url)
	}

	videos := getYoutubeVideos()

	pool, err := pb.StartPool()
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	var errs int

	for i, v := range videos {
		wg.Add(1)

		go func() {
			defer wg.Done()

			// bars should only be added when needed (otherwise all
			// bars will be rendered immediately)
			b := pb.Full.New(0).
				Set("prefix", v.Url).
				SetRefreshRate(time.Second).
				// SetTemplateString(BarTemplate)
				SetTemplateString(SpinnerTemplate)
			pool.Add(b)

			b.Start()
			// TODO: handle 403 / Sign in
			if err := v.Download(b); err != nil {
				b.SetCurrent(int64(100))
				b.SetTemplateString(`{{string . "prefix" }} ` + err.Error())
				// b.SetTemplateString("\r")
				b.Finish()
				errs++
				return
			}
			// change style when done
			b.SetCurrent(int64(100))
			b.SetTemplateString(`{{string . "prefix" }} done`)
			b.Finish()
			// wg.Done() // blocks eternally!
		}()

		// limit concurrent downloads
		if (i+1)%MaxConcurrent == 0 {
			wg.Wait()
		}

	}
	wg.Wait()
	if err := pool.Stop(); err != nil {
		panic(err)
	}

	fmt.Println(errs, "errors")

	// browseFiles(Cfg.Dest)
}
