package scraper

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/liweiyi88/gti/internal/database"
)

const ghTrendScrapePath = ".Box-row .h3.lh-condensed a[href]"
const ghTrendScrapeBaseURL = "https://github.com/trending"

type GhTrendScraper struct {
	url, path string
	db        database.DB
}

func NewGhTrendScraper() *GhTrendScraper {
	return &GhTrendScraper{
		url:  ghTrendScrapeBaseURL,
		path: ghTrendScrapePath,
		db:   database.GetInstance(),
	}
}

func (gh *GhTrendScraper) Scrape() {
	c := colly.NewCollector()

	repos := make([]string, 0)

	c.OnHTML(gh.path, func(e *colly.HTMLElement) {
		link := e.Attr("href")

		if strings.HasPrefix(link, "/") {
			link = strings.TrimLeft(link, "/")
		}

		repos = append(repos, link)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("scraping: %s \n", r.URL.String())
	})

	c.Visit(gh.url)

	// select today's trend and sort by rank

	for _, repo := range repos {

	}

	// save trends in DB
}
