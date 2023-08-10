package trending

import (
	"context"
	"fmt"

	"github.com/liweiyi88/gti/github"
	"github.com/liweiyi88/gti/model"
	"golang.org/x/sync/errgroup"
)

func FetchRepositories(ctx context.Context, grr *model.GhRepositoryRepo, trr *model.TrendingRepositoryRepo, gh *github.Client) error {
	unlinkedRepositories, err := trr.FindUnlinkedRepositories(ctx)

	if err != nil {
		return fmt.Errorf("failed to query unlinked repositories: %v", err)
	}

	repos, err := grr.FindRepositoriesByNames(ctx, unlinkedRepositories)

	if err != nil {
		return fmt.Errorf("failed to query repositories by names: %v", err)
	}

	reposNotExist := make([]string, 0)

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
			reposNotExist = append(reposNotExist, unlinkedRepo)
		}
	}

	group, ctx := errgroup.WithContext(ctx)

	for _, repo := range reposNotExist {
		repo := repo

		group.Go(func() error {
			repository, err := gh.GetRepository(ctx, repo)

			if err != nil {
				return err
			}

			lastInsertId, err := grr.Save(ctx, repository)
			repository.Id = int(lastInsertId)

			if err != nil {
				return err
			}

			return trr.LinkRepository(ctx, repository)
		})
	}

	if err := group.Wait(); err != nil {
		return fmt.Errorf("failed to fetch and save github repository details: %v", err)
	}

	return nil
}
