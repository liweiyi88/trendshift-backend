package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/liweiyi88/gti/config"
	"github.com/liweiyi88/gti/database"
	"github.com/liweiyi88/gti/github"
	"github.com/liweiyi88/gti/model"
	"github.com/liweiyi88/gti/model/opt"
	"github.com/liweiyi88/gti/utils/dbutils"
	"github.com/liweiyi88/gti/utils/sliceutils"
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
	Short: "Fetch latest repostiory details from GitHub and update db records",
	RunE: func(cmd *cobra.Command, args []string) error {
		config.Init()

		ctx, stop := context.WithCancel(context.Background())
		db := database.GetInstance(ctx)
		gh := github.NewClient(config.GitHubToken)

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

		repositoryRepo := model.NewGhRepositoryRepo(db, dbutils.NewQueryBuilder())

		repositories, err := repositoryRepo.FindAll(
			ctx,
			opt.Start(start),
			opt.End(end),
			opt.Limit(limit),
		)

		if err != nil {
			return fmt.Errorf("failed to find repositories: %v", err)
		}

		chulks := sliceutils.Chunk[model.GhRepository](repositories, 200)

		for _, chulk := range chulks {
			err := syncRepositories(ctx, chulk, *repositoryRepo, gh)

			if err != nil {
				return fmt.Errorf("could not sync repositories: %v", err)
			}

			slog.Info(fmt.Sprintf("completed batch update for %d repositories", len(chulk)))
		}

		slog.Info("repositories update completed.")
		return nil
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
