package scraper

import (
	"context"
	"testing"

	"github.com/liweiyi88/trendshift-backend/config"
	"github.com/liweiyi88/trendshift-backend/model"
	"golang.org/x/sync/errgroup"
)

func TestScrape(t *testing.T) {
	scraper := NewTrendingRepositoryScraper(&model.TrendingRepositoryRepo{})

	all, golang, php := scraper.getTrendPageUrl(""), scraper.getTrendPageUrl("Go"), scraper.getTrendPageUrl("PHP")

	expcts := []struct {
		actual any
		want   any
	}{
		{
			actual: all,
			want:   "https://github.com/trending",
		},
		{
			actual: golang,
			want:   "https://github.com/trending/Go?since=daily",
		},
		{
			actual: php,
			want:   "https://github.com/trending/PHP?since=daily",
		},
	}

	for _, test := range expcts {
		if test.actual != test.want {
			t.Errorf("expect: %v, actual got: %v", test.want, test.actual)
		}
	}

	group, _ := errgroup.WithContext(context.Background())

	for _, language := range config.LanguageToScrape {
		language := language
		group.Go(func() error {
			repositories := scraper.scrape(language)

			if len(repositories) == 0 {
				t.Logf("could not scrape trending repositories from GitHub, language: %s", language)
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		t.Error(err)
	}
}
