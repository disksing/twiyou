package cron

import (
	"fmt"
	"net/http"

	"github.com/disksing/twiyou/scraper"
)

// Vercel cron job entry point.

func Handle(w http.ResponseWriter, r *http.Request) {
	scraper, err := scraper.NewScraper()
	if err != nil {
		fmt.Fprintf(w, "failed to create scraper: %v", err)
		return
	}
	defer scraper.Close()
	err = scraper.Run()
	if err != nil {
		fmt.Fprintf(w, "failed to run scraper: %v", err)
		return
	}
	fmt.Fprintf(w, "ok")
}
