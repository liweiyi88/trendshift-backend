package scraper

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/liweiyi88/gti/trending"
)

const ghTrendScrapePath = ".Box-row .h3.lh-condensed a[href]"
const ghTrendScrapeBaseURL = "https://github.com/trending"

type GhTrendScraper struct {
	url, path string
	trendRepo *trend.TrendingRepositoryRepo
}

func NewGhTrendScraper(trendRepo *trend.TrendingRepositoryRepo) *GhTrendScraper {
	return &GhTrendScraper{
		url:       ghTrendScrapeBaseURL,
		path:      ghTrendScrapePath,
		trendRepo: trendRepo,
	}
}

func (gh *GhTrendScraper) Scrape(ctx context.Context, language string) error {
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
	rankedTrends, err := gh.trendRepo.FindRankedTrendsByDate(ctx, now, language)

	if err != nil {
		return fmt.Errorf("failed to retrieve ranked trending repositoris: %v", err)
	}

	for index, repo := range repos {
		rank := index + 1

		rankedTrend, ok := rankedTrends[rank]

		if ok && rankedTrend.RepoFullName != "" {
			// if trending repo exist, do update.
			rankedTrend.RepoFullName = repo
			rankedTrend.ScrapedAt, rankedTrend.TrendDate = now, now

			gh.trendRepo.Update(ctx, rankedTrend)
		} else {
			// trending repo does not exist, do insert.
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
			} else {
				trend.Language = sql.NullString{
					String: "",
					Valid:  false,
				}
			}

			gh.trendRepo.Save(ctx, trend)
		}
	}

	return nil
}

func (gh *GhTrendScraper) getTrendPageUrl(language string) string {
	language = strings.TrimSpace(language)

	if language != "" {
		return fmt.Sprintf("%s/%s?since=daily", gh.url, url.QueryEscape(language))
	}

	return gh.url
}