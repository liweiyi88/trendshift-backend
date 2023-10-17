package scraper

import (
	"context"
	"fmt"
	"testing"

	"github.com/liweiyi88/trendshift-backend/config"
	"github.com/liweiyi88/trendshift-backend/model"
	"golang.org/x/sync/errgroup"
)

func TestGetTrendPageUrl(t *testing.T) {
	scraper := NewTrendingDeveloperScraper(&model.TrendingDeveloperRepo{})

	all, golang, php := scraper.getTrendPageUrl(""), scraper.getTrendPageUrl("Go"), scraper.getTrendPageUrl("PHP")

	expcts := []struct {
		actual any
		want   any
	}{
		{
			actual: all,
			want:   "https://github.com/trending/developers",
		},
		{
			actual: golang,
			want:   "https://github.com/trending/developers/Go?since=daily",
		},
		{
			actual: php,
			want:   "https://github.com/trending/developers/PHP?since=daily",
		},
	}

	for _, test := range expcts {
		if test.actual != test.want {
			t.Errorf("expect: %v, actual got: %v", test.want, test.actual)
		}
	}

	group, ctx := errgroup.WithContext(context.Background())

	for _, language := range config.LanguageToScrape {
		group.Go(func() error {
			developers := scraper.scrape(ctx, language)

			if len(developers) == 0 {
				return fmt.Errorf("could not scrape trending developers from GitHub, language: %s", language)
			}

			return nil
		})

		if err := group.Wait(); err != nil {
			t.Error(err)
		}
	}
}
