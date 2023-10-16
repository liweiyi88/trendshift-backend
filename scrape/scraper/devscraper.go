package scraper

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/liweiyi88/gti/model"
	"github.com/liweiyi88/gti/utils/dbutils"
)

type TrendingDeveloperScraper struct {
	url, path             string
	trendingDeveloperRepo *model.TrendingDeveloperRepo
}

func NewTrendingDeveloperScraper(trendingDeveloperRepo *model.TrendingDeveloperRepo) *TrendingDeveloperScraper {
	return &TrendingDeveloperScraper{
		url:                   ghTrendScrapeBaseURL + "/developers",
		path:                  ghTrendScrapePath,
		trendingDeveloperRepo: trendingDeveloperRepo,
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
		// fmt.Printf("scraping: %s \n", r.URL.String())
	})

	c.Visit(ds.getTrendPageUrl(language))

	now := time.Now()

	rankedTrendingDevelopers, err := ds.trendingDeveloperRepo.FindRankedTrendingDevelopersByDate(ctx, now, language)

	if err != nil {
		return fmt.Errorf("failed to retrieve ranked trending developers: %v", err)
	}

	for index, developer := range developers {
		rank := index + 1

		trendingDeveloper, ok := rankedTrendingDevelopers[rank]

		if ok {
			// if trending developer exist, do update.
			trendingDeveloper.Username = developer
			trendingDeveloper.ScrapedAt, trendingDeveloper.TrendDate = now, now

			ds.trendingDeveloperRepo.Update(ctx, trendingDeveloper)
		} else {
			// trending developer does not exist, do insert.
			trendingDeveloper := model.TrendingDeveloper{
				Username:  developer,
				ScrapedAt: now,
				TrendDate: now,
				Rank:      rank,
			}

			if language != "" {
				trendingDeveloper.Language = dbutils.NullString{
					NullString: sql.NullString{String: strings.ToLower(language),
						Valid: true,
					},
				}
			} else {
				trendingDeveloper.Language = dbutils.NullString{
					NullString: sql.NullString{
						String: "",
						Valid:  false,
					}}
			}

			err = ds.trendingDeveloperRepo.Save(ctx, trendingDeveloper)
		}
	}

	return err
}

func (ds *TrendingDeveloperScraper) getTrendPageUrl(language string) string {
	language = strings.TrimSpace(language)

	if language != "" {
		return fmt.Sprintf("%s/%s?since=daily", ds.url, url.QueryEscape(language))
	}

	return ds.url
}
