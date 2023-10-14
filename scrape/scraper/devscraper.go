package scraper

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
)

type TrendingDeveloperScraper struct {
	url, path string
}

func NewTrendingDeveloperScraper() *TrendingDeveloperScraper {
	return &TrendingDeveloperScraper{
		url:  ghTrendScrapeBaseURL + "/developers",
		path: ghTrendScrapePath,
	}
}

func (ds *TrendingDeveloperScraper) GetType() string {
	return "developer"
}

func (ds *TrendingDeveloperScraper) Scrape(ctx context.Context, language string) error {
	c := colly.NewCollector()

	developers := make([]string, 0)

	c.OnHTML(ds.path, func(e *colly.HTMLElement) {
		link := e.Attr("href")

		if strings.HasPrefix(link, "/") {
			link = strings.TrimLeft(link, "/")
		}

		developers = append(developers, link)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("scraping: %s \n", r.URL.String())
	})

	c.Visit(ds.getTrendPageUrl(language))

	fmt.Printf("%+v", developers)

	// Save

	return nil
}

func (ds *TrendingDeveloperScraper) getTrendPageUrl(language string) string {
	language = strings.TrimSpace(language)

	if language != "" {
		return fmt.Sprintf("%s/%s?since=daily", ds.url, url.QueryEscape(language))
	}

	return ds.url
}
