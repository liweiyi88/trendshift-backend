package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"github.com/getsentry/sentry-go"
	"github.com/liweiyi88/trendshift-backend/config"
	"github.com/liweiyi88/trendshift-backend/database"
	"github.com/liweiyi88/trendshift-backend/github"
	"github.com/liweiyi88/trendshift-backend/model"
	"github.com/liweiyi88/trendshift-backend/model/opt"
	"github.com/liweiyi88/trendshift-backend/utils/dbutils"
	"github.com/liweiyi88/trendshift-backend/utils/sliceutils"
	"github.com/spf13/cobra"
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
	Short: "Fetch latest repostiory details from GitHub and update db records",
	Run: func(cmd *cobra.Command, args []string) {
		config.Init()

		ctx, stop := context.WithCancel(context.Background())
		db := database.GetInstance(ctx)
		gh := github.NewClient(config.GitHubToken)

		defer func() {
			err := db.Close()

			if err != nil {
				slog.Error("failed to close db", slog.Any("error", err))
				sentry.CaptureException(err)
			}

			stop()
			sentry.Flush(2 * time.Second)
		}()

		if start != "" {
			_, err := time.Parse("2006-01-02 15:04:05", start)
			if err != nil {
				slog.Error("failed to parse start time", slog.Any("error", err))
				sentry.CaptureException(err)
				return
			}
		}

		if end != "" {
			_, err := time.Parse("2006-01-02 15:04:05", end)
			if err != nil {
				slog.Error("failed to parse end time", slog.Any("error", err))
				sentry.CaptureException(err)
				return
			}
		}

		appSignal := make(chan os.Signal, 3)
		signal.Notify(appSignal, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-appSignal
			stop()
		}()

		repositoryRepo := model.NewGhRepositoryRepo(db, dbutils.NewQueryBuilder())

		repositories, err := repositoryRepo.FindAll(
			ctx,
			opt.Start(start),
			opt.End(end),
			opt.Limit(limit),
		)

		if err != nil {
			slog.Error("failed to find repositories", slog.Any("error", err))
			sentry.CaptureException(err)
			return
		}

		chulks := sliceutils.Chunk[model.GhRepository](repositories, 200)

		for _, chulk := range chulks {
			err := syncRepositories(ctx, chulk, *repositoryRepo, gh)

			if err != nil {
				slog.Error("could not sync repositories", slog.Any("error", err))
				sentry.CaptureException(err)
				return
			}

			slog.Info(fmt.Sprintf("completed batch update for %d repositories", len(chulk)))
		}

		slog.Info("repositories update completed.")
	},
}

func syncRepositories(
	ctx context.Context,
	repositories []model.GhRepository,
	repositoryRepo model.GhRepositoryRepo,
	gh *github.Client,
) error {
	group, ctx := errgroup.WithContext(ctx)

	// Follow the github best practice to avoid reaching secondary rate limit
	// see https://docs.github.com/en/rest/guides/best-practices-for-using-the-rest-api?apiVersion=2022-11-28#dealing-with-secondary-rate-limits
	limiter := time.Tick(20 * time.Millisecond)

	for _, repository := range repositories {
		<-limiter

		repository := repository

		group.Go(func() error {
			ghRepository, err := gh.GetRepository(ctx, repository.FullName)

			if err != nil {
				if errors.Is(err, github.ErrNotFound) {
					slog.Info(fmt.Sprintf("not found on GitHub, repository: %s", repository.FullName))
				} else if errors.Is(err, github.ErrAccessBlocked) {
					slog.Info(fmt.Sprintf("repository access blocked due to leagl reason, repository: %s", repository.FullName))
				} else {
					return fmt.Errorf("failed to get repository details from GitHub: %v", err)
				}
			}

			repository.Description = ghRepository.Description
			repository.Forks = ghRepository.Forks
			repository.Stars = ghRepository.Stars
			repository.Owner = ghRepository.Owner
			repository.Language = ghRepository.Language // Language can also be updated
			repository.DefaultBranch = ghRepository.DefaultBranch

			return repositoryRepo.Update(ctx, repository)
		})
	}

	return group.Wait()
}
