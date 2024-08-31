package github

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/liweiyi88/trendshift-backend/database"
	"github.com/liweiyi88/trendshift-backend/model"
	"github.com/liweiyi88/trendshift-backend/utils/sliceutils"
	"golang.org/x/sync/errgroup"
)

const chulkSize = 200

type SyncHandler struct {
	db             database.DB
	repositoryRepo *model.GhRepositoryRepo
	developerRepo  *model.DeveloperRepo
	client         *Client
}

func NewSyncHandler(db database.DB, repositoryRepo *model.GhRepositoryRepo, developerRepo *model.DeveloperRepo, client *Client) *SyncHandler {
	return &SyncHandler{
		db, repositoryRepo, developerRepo, client,
	}
}

func (s *SyncHandler) updateRepositories(ctx context.Context, repositories []model.GhRepository) error {
	group, ctx := errgroup.WithContext(ctx)

	// Follow the github best practice to avoid reaching secondary rate limit
	// see https://docs.github.com/en/rest/guides/best-practices-for-using-the-rest-api?apiVersion=2022-11-28#dealing-with-secondary-rate-limits
	limiter := time.NewTicker(20 * time.Millisecond)
	defer limiter.Stop()

	for _, repository := range repositories {
		<-limiter.C

		repository := repository

		group.Go(func() error {
			ghRepository, err := s.client.GetRepository(ctx, repository.FullName)

			if err != nil {
				if errors.Is(err, ErrNotFound) {
					slog.Info(fmt.Sprintf("repository not found on GitHub, repository: %s", repository.FullName))
				} else if errors.Is(err, ErrAccessBlocked) {
					slog.Info(fmt.Sprintf("repository access blocked, repository: %s", repository.FullName))
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
			repository.Homepage = ghRepository.Homepage

			return s.repositoryRepo.Update(ctx, repository)
		})
	}

	return group.Wait()
}

func (s *SyncHandler) updateDevelopers(ctx context.Context, developers []model.Developer) error {
	group, ctx := errgroup.WithContext(ctx)

	// Follow the github best practice to avoid reaching secondary rate limit
	// see https://docs.github.com/en/rest/guides/best-practices-for-using-the-rest-api?apiVersion=2022-11-28#dealing-with-secondary-rate-limits
	limiter := time.NewTicker(20 * time.Millisecond)
	defer limiter.Stop()

	for _, developer := range developers {
		<-limiter.C

		developer := developer

		group.Go(func() error {
			ghDeveloper, err := s.client.GetDeveloper(ctx, developer.Username)

			if err != nil {
				if errors.Is(err, ErrNotFound) {
					slog.Info(fmt.Sprintf("not found on GitHub, developer: %s", developer.Username))
				} else if errors.Is(err, ErrAccessBlocked) {
					slog.Info(fmt.Sprintf("developer access blocked due to leagl reason, developer: %s", developer.Username))
				} else {
					return fmt.Errorf("failed to get developer details from GitHub: %v", err)
				}
			}

			developer.AvatarUrl = ghDeveloper.AvatarUrl
			developer.Name = ghDeveloper.Name
			developer.Company = ghDeveloper.Company
			developer.Blog = ghDeveloper.Blog
			developer.Location = ghDeveloper.Location
			developer.Email = ghDeveloper.Email
			developer.Bio = ghDeveloper.Bio
			developer.TwitterUsername = ghDeveloper.TwitterUsername
			developer.PublicRepos = ghDeveloper.PublicRepos
			developer.PublicGists = ghDeveloper.PublicGists
			developer.Followers = ghDeveloper.Followers
			developer.Following = ghDeveloper.Following

			return s.developerRepo.Update(ctx, developer)
		})
	}

	return group.Wait()
}

func (s *SyncHandler) syncRepositories(ctx context.Context, opts ...any) error {
	repositories, err := s.repositoryRepo.FindAll(
		ctx,
		opts...,
	)

	if err != nil {
		return fmt.Errorf("failed to find repositories: %v", err)
	}

	chulks := sliceutils.Chunk[model.GhRepository](repositories, chulkSize)

	for _, chulk := range chulks {
		err := s.updateRepositories(ctx, chulk)

		if err != nil {
			return fmt.Errorf("could not sync repositories: %v", err)
		}

		slog.Info(fmt.Sprintf("completed batch update for %d repositories", len(chulk)))
	}

	slog.Info("repositories update completed.")
	return nil
}

func (s *SyncHandler) syncDevelopers(ctx context.Context, opts ...any) error {
	developers, err := s.developerRepo.FindAll(ctx, opts...)

	if err != nil {
		return fmt.Errorf("failed to find developers: %v", err)
	}

	chulks := sliceutils.Chunk[model.Developer](developers, chulkSize)

	for _, chulk := range chulks {
		err := s.updateDevelopers(ctx, chulk)

		if err != nil {
			return fmt.Errorf("could not sync developers: %v", err)
		}

		slog.Info(fmt.Sprintf("completed batch update for %d developers", len(chulk)))
	}

	slog.Info("developers update completed.")
	return nil
}

func (s *SyncHandler) Handle(ctx context.Context, action string, opts ...any) error {
	switch action {
	case "repository":
		return s.syncRepositories(ctx, opts...)
	case "developer":
		return s.syncDevelopers(ctx, opts...)
	default:
		return errors.New("invalid search action")
	}
}
