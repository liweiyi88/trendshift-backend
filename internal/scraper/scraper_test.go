package scraper

import "testing"

func TestScrape(t *testing.T) {
	Scrape("https://github.com/trending")
}
