package scraper

import (
	"fmt"

	"github.com/gocolly/colly/v2"
)

func Scrape(url string) {
	c := colly.NewCollector()

	// On every a element which has href attribute call callback
	c.OnHTML(".Box-row .h3.lh-condensed a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// Print link
		fmt.Printf("Link: %s\n", link)
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit(url)
}
