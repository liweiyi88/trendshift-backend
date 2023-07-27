package scraper

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/liweiyi88/gti/internal/database"
	"github.com/liweiyi88/gti/internal/trend"
)

const ghTrendScrapePath = ".Box-row .h3.lh-condensed a[href]"
const ghTrendScrapeBaseURL = "https://github.com/trending"

type GhTrendScraper struct {
	url, path string
	db        database.DB
	trendRepo *trend.TrendingRepositoryRepo
}

func NewGhTrendScraper(trendRepo *trend.TrendingRepositoryRepo) *GhTrendScraper {
	return &GhTrendScraper{
		url:       ghTrendScrapeBaseURL,
		path:      ghTrendScrapePath,
		db:        database.GetInstance(),
		trendRepo: trendRepo,
	}
}

func (gh *GhTrendScraper) Scrape(language string) error {
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

	c.Visit(gh.getTrendPageUrl(language))

	now := time.Now()
	rankedTrends, err := gh.trendRepo.FindRankedTrendsByDate(context.Background(), now)
	if err != nil {
		return err
	}

	for index, repo := range repos {
		rank := index + 1

		if rankedTrend, ok := rankedTrends[rank]; ok && rankedTrend.RepoFullName != "" {
			// if trend exist， do update.
			rankedTrend.RepoFullName = repo
			rankedTrend.ScrapedAt, rankedTrend.TrendDate = now, now

			gh.trendRepo.Update(context.Background(), rankedTrend)
		} else {
			// trend does not exist, do insert.
			trend := trend.TrendingRepository{
				RepoFullName: repo,
				ScrapedAt:    now,
				TrendDate:    now,
				Rank:         rank,
			}

			if language != "" {
				trend.Language = sql.NullString{
					String: strings.ToLower(language),
					Valid:  true,
				}
			}

			gh.trendRepo.Save(context.Background(), trend)
		}
	}

	return nil
}

func (gh *GhTrendScraper) getTrendPageUrl(language string) string {
	language = strings.TrimSpace(language)

	if language != "" {
		return fmt.Sprintf("%s/%s", gh.url, language)
	}

	return gh.url
}
