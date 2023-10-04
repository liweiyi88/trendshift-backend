package trending

import (
	"context"
	"fmt"

	"github.com/liweiyi88/gti/github"
	"github.com/liweiyi88/gti/model"
	"github.com/liweiyi88/gti/search"
	"golang.org/x/sync/errgroup"
)

func FetchRepositories(
	ctx context.Context,
	grr *model.GhRepositoryRepo,
	trr *model.TrendingRepositoryRepo,
	gh *github.Client,
	search search.Search) error {
	unlinkedRepositories, err := trr.FindUnlinkedRepositories(ctx)

	if err != nil {
		return fmt.Errorf("failed to query unlinked repositories: %v", err)
	}

	repos, err := grr.FindRepositoriesByNames(ctx, unlinkedRepositories)

	if err != nil {
		return fmt.Errorf("failed to query repositories by names: %v", err)
	}

	repoNamesNotExist := make([]string, 0)

	// if repository exist in DB, then we update the relationship
	for _, unlinkedRepo := range unlinkedRepositories {
		exist := false

		for _, repo := range repos {
			if repo.FullName == unlinkedRepo {
				exist = true
				err := trr.LinkRepository(ctx, repo)

				if err != nil {
					return err
				}
			}
		}

		if !exist {
			repoNamesNotExist = append(repoNamesNotExist, unlinkedRepo)
		}
	}

	group, ctx := errgroup.WithContext(ctx)

	repositoriesNotExist := make([]model.GhRepository, 0)

	for _, repo := range repoNamesNotExist {
		repo := repo

		group.Go(func() error {
			repository, err := gh.GetRepository(ctx, repo)

			if err != nil {
				return err
			}

			lastInsertId, err := grr.Save(ctx, repository)
			repository.Id = int(lastInsertId)

			if err != nil {
				return fmt.Errorf("failed to save repository: %v", err)
			}

			err = trr.LinkRepository(ctx, repository)

			if err != nil {
				return fmt.Errorf("failed to link repository: %v", err)
			}

			repositoriesNotExist = append(repositoriesNotExist, repository)
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return fmt.Errorf("failed to fetch and save github repository details: %v", err)
	}

	return search.UpsertRepositories(repositoriesNotExist...)
}
