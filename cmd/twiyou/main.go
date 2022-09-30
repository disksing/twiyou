package main

import (
	"log"

	"github.com/disksing/twiyou"
)

func main() {
	scraper, err := twiyou.NewScraper()
	if err != nil {
		log.Fatal(err)
	}
	defer scraper.Close()
	err = scraper.Run()
	if err != nil {
		log.Fatal(err)
	}
}
