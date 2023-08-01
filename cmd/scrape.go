package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/liweiyi88/gti/config"
	"github.com/liweiyi88/gti/database"
	"github.com/liweiyi88/gti/github"
	"github.com/liweiyi88/gti/global"
	"github.com/liweiyi88/gti/scraper"
	"github.com/liweiyi88/gti/trendingsvc"
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

		repositories := global.InitRepositories(db)
		scraper := scraper.NewGhTrendScraper(repositories.TrendingRepositoryRepo)

		languageToScrape := []string{"", "javascript", "python", "go", "java", "php", "c++", "c", "typescript", "ruby", "c#", "rust"}

		group, groupCtx := errgroup.WithContext(ctx)

		slog.Info(fmt.Sprintf("scraping for languages: %s...", strings.Join(languageToScrape, ",")))

		for _, language := range languageToScrape {
			language := language
			group.Go(func() error {
				return scraper.Scrape(groupCtx, language)
			})
		}

		if err := group.Wait(); err != nil {
			log.Fatalf("failed to scrape trending page: %v", err)
		}

		slog.Info("linking repositories...")
		err := trendingsvc.LinkRepositories(ctx, repositories.GhRepositoryRepo, repositories.TrendingRepositoryRepo, github.NewClient(config.GitHubToken))

		if err != nil {
			log.Fatalf("failed to link repositories trending page: %v", err)
		}

		slog.Info("scrape completed.")
	},
}
