package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/liweiyi88/gti/config"
	"github.com/liweiyi88/gti/database"
	"github.com/liweiyi88/gti/repository"
	"github.com/liweiyi88/gti/scraper"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
)

func init() {
	rootCmd.AddCommand(scrapeCmd)
}

var scrapeCmd = &cobra.Command{
	Use:   "scrape",
	Short: "Scrape the trending repositories form GitHub trending page.",
	Run: func(cmd *cobra.Command, args []string) {
		config.Init()

		ctx, stop := context.WithCancel(context.Background())
		db := database.GetInstance(ctx)

		defer func() {
			err := db.Close()

			if err != nil {
				slog.Error("failed to close db: %v", err)
			}

			stop()
		}()

		appSignal := make(chan os.Signal, 3)
		signal.Notify(appSignal, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-appSignal
			stop()
		}()

		repositories := repository.InitRepositories(db)
		scraper := scraper.NewGhTrendScraper(repositories.TrendingRepositoryRepo)

		languageToScrape := []string{"", "javascript", "python", "Go", "java", "php", "c++", "c", "typescript", "ruby", "c#", "rust"}

		group, ctx := errgroup.WithContext(ctx)

		for _, language := range languageToScrape {
			language := language
			group.Go(func() error {
				return scraper.Scrape(ctx, language)
			})
		}

		if err := group.Wait(); err != nil {
			log.Fatalf("failed to scrape trending page: %v", err)
		}
	},
}
