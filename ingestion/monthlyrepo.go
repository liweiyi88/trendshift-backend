package ingestion

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/liweiyi88/trendshift-backend/github"
	"github.com/liweiyi88/trendshift-backend/model"
	"github.com/liweiyi88/trendshift-backend/utils/datetime"
	"github.com/liweiyi88/trendshift-backend/utils/dbutils"
	"golang.org/x/sync/errgroup"
)

const batchSize = 1000

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

func (ingestor *MonthlyRepoDataIngestor) ingest(ctx context.Context, start, end time.Time, insight model.RepositoryMonthlyInsightWithName) error {
	repoName := insight.RepositoryName

	var forks, stars, mergedPrs, issues, closedIssues int
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
		all, closed, err := ingestor.fetchIssues(gctx, repoName, start, end)
		if err != nil {
			slog.Error("failed to fetch monthly issues",
				slog.String("repository", repoName),
				slog.Any("error", err))

			return err
		}
		issues = all
		closedIssues = closed
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	insight.Stars = dbutils.NewNullInt64(stars)
	insight.Forks = dbutils.NewNullInt64(forks)
	insight.Issues = dbutils.NewNullInt64(issues)
	insight.ClosedIssues = dbutils.NewNullInt64(closedIssues)
	insight.MergedPrs = dbutils.NewNullInt64(mergedPrs)

	now := time.Now()
	if insight.Month == int(now.Month())-1 {
		insight.CompletedAt = dbutils.NewNullTime(now)
	}

	insight.LastIngestedAt = dbutils.NewNullTime(now)

	slog.Debug("completed fetching repo monthly data",
		slog.String("repository", repoName),
		slog.Int("stars", stars),
		slog.Int("forks", forks),
		slog.Int("issues", issues),
		slog.Int("closedIssues", closedIssues),
		slog.Int("mergedPrs", mergedPrs))

	return ingestor.rmr.Update(ctx, insight)
}

func (ingestor *MonthlyRepoDataIngestor) Ingest(ctx context.Context, month, year int) (bool, error) {
	_, err := ingestor.rmr.CreateMonthlyInsightsIfNotExist(ctx, month, year)

	if err != nil {
		return false, fmt.Errorf("failed to create monthly insights if not exist, error: %w", err)
	}

	montlyRepoInsights, err := ingestor.rmr.FindIncompletedLastIngestedBefore(ctx, datetime.StartOfToday(), batchSize)
	if err != nil {
		return false, fmt.Errorf("failed to ingest repository monthly data, error: %w", err)
	}

	if len(montlyRepoInsights) == 0 {
		return true, nil
	}

	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	end := datetime.EndOfMonth(start)

	chunks := slices.Chunk(montlyRepoInsights, 10)

	for chunk := range chunks {
		g, gctx := errgroup.WithContext(ctx)
		for _, insight := range chunk {
			g.Go(func() error {
				return ingestor.ingest(gctx, start, end, insight)
			})
		}

		if err := g.Wait(); err != nil {
			return false, err
		}

		ingestor.gh.TokenPool.ToString()
	}

	return false, nil
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
			return nil, fmt.Errorf("failed to fetch paginated data: %w", err)
		}

		all = append(all, data...)

		if nextCursor == nil {
			break
		}
		cursor = nextCursor
	}

	return all, nil
}

func (ingestor *MonthlyRepoDataIngestor) fetchIssues(ctx context.Context, repository string, start, end time.Time) (int, int, error) {
	data, err := fetchPaginated(ctx, repository, start, end, ingestor.gh.GetIssues)
	closed := 0

	for _, v := range data {
		if v.Closed && !v.ClosedAt.IsZero() {
			closed++
		}
	}

	return len(data), closed, err
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
