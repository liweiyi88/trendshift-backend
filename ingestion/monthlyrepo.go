package ingestion

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/liweiyi88/trendshift-backend/github"
	"github.com/liweiyi88/trendshift-backend/model"
	"github.com/liweiyi88/trendshift-backend/utils/datetime"
	"golang.org/x/sync/errgroup"
)

// fetch from database, where completed_at is null
// fetch monthly repository data from GitHub, upsert database, if the month is last month, then set the completed_at

// 1. CreateMonthlyInsightsIfNotExist
// 2. select 100 repositories from the DB where updated_at is not today,
// use updated_at ASC, fetch data and update the record (set completed_at if it is past month)
// 3. keep select 100 repositories until null, then we can return
type MonthlyRepoDataIngestor struct {
	gh  *github.Client
	rmr *model.RepositoryMonthlyInsightRepo
}

func NewMonthlyRepoDataIngestor(rmr *model.RepositoryMonthlyInsightRepo, gh *github.Client) *MonthlyRepoDataIngestor {
	return &MonthlyRepoDataIngestor{
		rmr: rmr,
		gh:  gh,
	}
}

func (ingestor *MonthlyRepoDataIngestor) Ingest(ctx context.Context, month, year int) error {
	_, err := ingestor.rmr.CreateMonthlyInsightsIfNotExist(ctx, month, year)

	if err != nil {
		return fmt.Errorf("failed to create monthly insights if not exist, error: %v", err)
	}

	montlyRepoInsights, err := ingestor.rmr.FindIncompletedLastIngestedBefore(ctx, datetime.StartOfToday(), 1000)
	if err != nil {
		return fmt.Errorf("failed to ingest repository monthly data, error: %v", err)
	}

	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	end := datetime.EndOfMonth(start)

	for _, insight := range montlyRepoInsights {
		fmt.Print(insight.Id)
		// repoName := insight.RepositoryName

		repoName := "liweiyi88/onedump"

		var forks, stars, mergedPrs, closedIssues int
		g, gctx := errgroup.WithContext(ctx)

		g.Go(func() error {
			count, err := ingestor.fetchForks(gctx, repoName, start, end)
			if err != nil {
				slog.Error("failed to fetch monthly forks",
					slog.String("repository", repoName),
					slog.Any("error", err))
				return err
			}
			forks = count
			return nil
		})

		g.Go(func() error {
			count, err := ingestor.fetchStars(gctx, repoName, start, end)
			if err != nil {
				slog.Error("failed to fetch monthly stars",
					slog.String("repository", repoName),
					slog.Any("error", err))
				return err
			}
			stars = count
			return nil
		})

		g.Go(func() error {
			count, err := ingestor.fetchMergedPrs(gctx, repoName, start, end)
			if err != nil {
				slog.Error("failed to fetch monthly merged prs",
					slog.String("repository", repoName),
					slog.Any("error", err))
				return err
			}
			mergedPrs = count
			return nil
		})

		g.Go(func() error {
			count, err := ingestor.fetchClosedIssues(gctx, repoName, start, end)
			if err != nil {
				slog.Error("failed to fetch monthly closed issues",
					slog.String("repository", repoName),
					slog.Any("error", err))
				return err
			}
			closedIssues = count
			return nil
		})

		if err := g.Wait(); err != nil {
			slog.Warn("skipping repo due to fetch error", slog.String("repository", repoName), slog.Any("error", err))
			continue
		}

		slog.Info("fetched repo monthly data",
			slog.String("repository", repoName),
			slog.Int("stars", stars),
			slog.Int("forks", forks),
			slog.Int("closedIssues", closedIssues),
			slog.Int("mergedPrs", mergedPrs))

		// forks, err := ingestor.fetchForks(ctx, insight.RepositoryName, start, end)
		// if err != nil {
		// 	// @todo handle rate limit issue
		// 	slog.Error("failed to fetch monthly forks", slog.String("repository", insight.RepositoryName), slog.Any("error", err))
		// 	continue
		// }

		// slog.Info("fetched repo monthly forks", slog.Int("forks", forks), slog.String("repository", insight.RepositoryName))

		// stars, err := ingestor.fetchStars(ctx, insight.RepositoryName, start, end)
		// if err != nil {
		// 	// @todo handle rate limit issue
		// 	slog.Error("failed to fetch monthly stars", slog.String("repository", insight.RepositoryName), slog.Any("error", err))
		// 	continue
		// }

		// slog.Info("fetched repo monthly stars", slog.Int("stars", stars), slog.String("repository", insight.RepositoryName))

		// return nil
		// fetch stars, fetch forks, fetch prs, fetch issues

		return nil
	}

	return nil
}

func fetchPaginated[T any](
	ctx context.Context,
	repository string,
	start, end time.Time,
	fetchPage func(ctx context.Context, owner, repo string, cursor *string, start, end *time.Time) ([]T, *string, error),
) ([]T, error) {
	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository name: %s", repository)
	}

	owner, repo := parts[0], parts[1]

	var cursor *string
	var all []T

	for {
		data, nextCursor, err := fetchPage(ctx, owner, repo, cursor, &start, &end)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch paginated data: %v", err)
		}

		all = append(all, data...)

		if nextCursor == nil {
			break
		}
		cursor = nextCursor
	}

	return all, nil
}

func (ingestor *MonthlyRepoDataIngestor) fetchClosedIssues(ctx context.Context, repository string, start, end time.Time) (int, error) {
	data, err := fetchPaginated(ctx, repository, start, end, ingestor.gh.GetClosedIssues)
	return len(data), err
}

func (ingestor *MonthlyRepoDataIngestor) fetchMergedPrs(ctx context.Context, repository string, start, end time.Time) (int, error) {
	data, err := fetchPaginated(ctx, repository, start, end, ingestor.gh.GetMergedPrs)
	return len(data), err
}

func (ingestor *MonthlyRepoDataIngestor) fetchForks(ctx context.Context, repository string, start, end time.Time) (int, error) {
	data, err := fetchPaginated(ctx, repository, start, end, ingestor.gh.GetRepositoryForks)
	return len(data), err
}

func (ingestor *MonthlyRepoDataIngestor) fetchStars(ctx context.Context, repository string, start, end time.Time) (int, error) {
	data, err := fetchPaginated(ctx, repository, start, end, ingestor.gh.GetRepositoryStars)
	return len(data), err
}
