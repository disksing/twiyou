package main

import (
	"log"

	"github.com/disksing/twiyou/scraper"
)

func main() {
	scraper, err := scraper.NewScraper()
	if err != nil {
		log.Fatal(err)
	}
	defer scraper.Close()
	err = scraper.Run()
	if err != nil {
		log.Fatal(err)
	}
}
