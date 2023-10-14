package scrape

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"log/slog"

	"github.com/liweiyi88/gti/config"
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
	repositories *global.Repositories
	search       search.Search
}

func NewScrapeHandler(repositories *global.Repositories, search search.Search) *ScrapeHandler {
	return &ScrapeHandler{
		repositories, search,
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

func (s *ScrapeHandler) saveTrendingRepositories(ctx context.Context) error {
	scraper := scraper.NewTrendingRepositoryScraper(s.repositories.TrendingRepositoryRepo)

	err := save(scraper, ctx)

	if err != nil {
		return err
	}

	slog.Info("linking repositories...")

	err = trending.FetchRepositories(
		ctx,
		s.repositories.GhRepositoryRepo,
		s.repositories.TrendingRepositoryRepo,
		github.NewClient(config.GitHubToken),
		s.search)

	if err != nil {
		log.Fatalf("failed to link repositories trending page: %v", err)
	}

	slog.Info("scrape completed.")
	return nil
}

func (s *ScrapeHandler) saveTrendingDevelopers(ctx context.Context) error {
	scraper := scraper.NewTrendingDeveloperScraper()

	err := save(scraper, ctx)

	if err != nil {
		return err
	}

	// TODO link

	slog.Info("scrape completed.")
	return nil
}
