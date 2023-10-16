package trending

import (
	"context"
	"fmt"

	"github.com/liweiyi88/trendshift-backend/github"
	"github.com/liweiyi88/trendshift-backend/global"
	"github.com/liweiyi88/trendshift-backend/model"
	"github.com/liweiyi88/trendshift-backend/search"
	"golang.org/x/sync/errgroup"
)

type GithubFetcher struct {
	gh           *github.Client
	search       search.Search
	repositories global.Repositories
}

func NewGithubFetcher(gh *github.Client, search search.Search, repositories global.Repositories) *GithubFetcher {
	return &GithubFetcher{
		gh, search, repositories,
	}
}

func (fetcher *GithubFetcher) FetchDevelopers(ctx context.Context) error {
	tdr, dr := fetcher.repositories.TrendingDeveloperRepo, fetcher.repositories.DeveloperRepo

	unlinkedDevelopers, err := tdr.FindUnlinkedDevelopers(ctx)

	if err != nil {
		return fmt.Errorf("failed to query unlinked developers: %v", err)
	}

	developers, err := dr.FindDevelopersByUsernames(ctx, unlinkedDevelopers)

	if err != nil {
		return fmt.Errorf("failed to query developers by names: %v", err)
	}

	devNamesNotExist := make([]string, 0)

	// if developer exist in DB, then we update the relationship
	for _, unlinkedDeveloper := range unlinkedDevelopers {
		exist := false

		for _, developer := range developers {
			if developer.Username == unlinkedDeveloper {
				exist = true

				err := tdr.LinkDeveloper(ctx, developer)

				if err != nil {
					return err
				}
			}
		}

		if !exist {
			devNamesNotExist = append(devNamesNotExist, unlinkedDeveloper)
		}
	}

	group, ctx := errgroup.WithContext(ctx)

	developersNotExist := make([]model.Developer, 0)

	for _, dev := range devNamesNotExist {
		devName := dev

		group.Go(func() error {
			developer, err := fetcher.gh.GetDeveloper(ctx, devName)

			if err != nil {
				return err
			}

			lastInsertId, err := dr.Save(ctx, developer)
			developer.Id = int(lastInsertId)

			if err != nil {
				return fmt.Errorf("failed to save developer: %v", err)
			}

			err = tdr.LinkDeveloper(ctx, developer)

			if err != nil {
				return fmt.Errorf("failed to link developer: %v", err)
			}

			developersNotExist = append(developersNotExist, developer)
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return fmt.Errorf("failed to fetch and save github developer details: %v", err)
	}

	return fetcher.search.UpsertDevelopers(developersNotExist...)
}

// Fetch repositories details from github rest api and save the relationship between trending_repositories and repositories.
func (fetcher *GithubFetcher) FetchRepositories(ctx context.Context) error {
	trr, grr := fetcher.repositories.TrendingRepositoryRepo, fetcher.repositories.GhRepositoryRepo

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
			repository, err := fetcher.gh.GetRepository(ctx, repo)

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

	return fetcher.search.UpsertRepositories(repositoriesNotExist...)
}
