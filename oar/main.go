package main

import (
	"fmt"
	"sync"
)

func main() {
	config := LoadConfig()
	labels := getBandcampLabels(config.Bandcamp.Username)

	var wg sync.WaitGroup
	releases := []BandcampRelease{}
	for i, label := range labels {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := label.getReleases(7)
			if len(r) == 0 {
				return
			}
			fmt.Println(i, label.UrlHints.Subdomain, len(r))
			releases = append(releases, r...)
		}()
	}
	wg.Wait()
	fmt.Printf("found %d releases from %d labels\n", len(releases), len(labels))
}
