package scrape

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"log/slog"

	"github.com/liweiyi88/gti/github"
	"github.com/liweiyi88/gti/global"
	"github.com/liweiyi88/gti/scrape/scraper"
	"github.com/liweiyi88/gti/search"
	"github.com/liweiyi88/gti/trending"
	"golang.org/x/sync/errgroup"
)

const (
	repository = "repository"
	developer  = "developer"
)

var languageToScrape = []string{"", "javascript", "python", "go", "java", "php", "c++", "c", "typescript", "ruby", "c#", "rust"}

type Scraper interface {
	Scrape(ctx context.Context, language string) error
	GetType() string
}

type ScrapeHandler struct {
	repositories  *global.Repositories
	search        search.Search
	githubFetcher *trending.GithubFetcher
}

func NewScrapeHandler(repositories *global.Repositories, search search.Search, gh *github.Client) *ScrapeHandler {
	return &ScrapeHandler{
		repositories:  repositories,
		search:        search,
		githubFetcher: trending.NewGithubFetcher(gh, search, *repositories),
	}
}

func (s *ScrapeHandler) Handle(ctx context.Context, action string) error {
	switch action {
	case repository:
		return s.saveTrendingRepositories(ctx)
	case developer:
		return s.saveTrendingDevelopers(ctx)
	default:
		return errors.New("invalid search action")
	}
}

func (s *ScrapeHandler) saveTrendingRepositories(ctx context.Context) error {
	scraper := scraper.NewTrendingRepositoryScraper(s.repositories.TrendingRepositoryRepo)

	err := save(scraper, ctx)

	if err != nil {
		return err
	}

	slog.Info("linking repositories...")

	err = s.githubFetcher.FetchRepositories(ctx)

	if err != nil {
		return fmt.Errorf("failed to fetch and link repositories from trending page: %v", err)
	}

	slog.Info("scrape completed.")
	return nil
}

func (s *ScrapeHandler) saveTrendingDevelopers(ctx context.Context) error {
	scraper := scraper.NewTrendingDeveloperScraper(s.repositories.TrendingDeveloperRepo)

	err := save(scraper, ctx)

	if err != nil {
		return err
	}

	slog.Info("linking developers...")

	err = s.githubFetcher.FetchDevelopers(ctx)

	if err != nil {
		return fmt.Errorf("failed to fetch and link developers from trending page: %v", err)
	}

	slog.Info("scrape completed.")
	return nil
}

// Scrape repositories or developers rank from GitHub Trending page and save them in DB.
func save(scraper Scraper, ctx context.Context) error {
	group, groupCtx := errgroup.WithContext(ctx)

	slog.Info(fmt.Sprintf("Scraping %s for languages: %s...", scraper.GetType(), strings.Join(languageToScrape, ",")))

	for _, language := range languageToScrape {
		language := language
		group.Go(func() error {
			return scraper.Scrape(groupCtx, language)
		})
	}

	if err := group.Wait(); err != nil {
		return fmt.Errorf("failed to scrape and save trending %s: %v", scraper.GetType(), err)
	}

	return nil
}
