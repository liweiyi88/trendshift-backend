package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/liweiyi88/gti/config"
	"github.com/liweiyi88/gti/database"
	"github.com/liweiyi88/gti/github"
	"github.com/liweiyi88/gti/model"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
)

var start string
var end string
var limit int

func init() {
	rootCmd.AddCommand(repositoryCmd)

	repositoryCmd.Flags().StringVarP(&start, "start", "s", "", "start")
	repositoryCmd.Flags().StringVarP(&end, "end", "e", "", "end")
	repositoryCmd.Flags().IntVarP(&limit, "limit", "l", 0, "limit")
}

var repositoryCmd = &cobra.Command{
	Use:   "repository",
	Short: "Fetch latest repostiory details from GitHub",
	Run: func(cmd *cobra.Command, args []string) {
		config.Init()

		ctx, stop := context.WithCancel(context.Background())
		db := database.GetInstance(ctx)

		if start != "" {
			_, err := time.Parse("2006-01-02 15:04:05", start)
			if err != nil {
				log.Fatalf("failed to parse start time: %v", err)
			}
		}

		if end != "" {
			_, err := time.Parse("2006-01-02 15:04:05", end)
			if err != nil {
				log.Fatalf("failed to parse end time: %v", err)
			}
		}

		defer func() {
			err := db.Close()

			if err != nil {
				slog.Error("failed to close db", slog.Any("error", err))
			}

			stop()
		}()

		appSignal := make(chan os.Signal, 3)
		signal.Notify(appSignal, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-appSignal
			stop()
		}()

		repositoryRepo := model.NewGhRepositoryRepo(db)

		var repositories []model.GhRepository
		var err error

		repositories, err = repositoryRepo.FindAll(ctx, start, end, limit)

		if err != nil {
			slog.Error("could not retrieve repositories", slog.Any("error", err))
			return
		}

		group, ctx := errgroup.WithContext(ctx)

		gh := github.NewClient(config.GitHubToken)

		requests := make(chan model.GhRepository, len(repositories))

		for _, repository := range repositories {
			requests <- repository
		}

		close(requests)

		// Follow the github best practice to avoid reaching secondary rate limit
		// see https://docs.github.com/en/rest/guides/best-practices-for-using-the-rest-api?apiVersion=2022-11-28#dealing-with-secondary-rate-limits
		limiter := time.Tick(20 * time.Millisecond)

		for repository := range requests {
			<-limiter
			repository := repository

			group.Go(func() error {
				ghRepository, err := gh.GetRepository(ctx, repository.FullName)

				if err != nil {
					return err
				}

				repository.Description = ghRepository.Description
				repository.Forks = ghRepository.Forks
				repository.Stars = ghRepository.Stars
				repository.Owner = ghRepository.Owner
				repository.DefaultBranch = ghRepository.DefaultBranch

				return repositoryRepo.Update(ctx, repository)
			})
		}

		if err := group.Wait(); err != nil {
			slog.Error("failed to fetch and save github repository details", slog.Any("error", err))
		}

		slog.Info("repositories update completed.")
	},
}
