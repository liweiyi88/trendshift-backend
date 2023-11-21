package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/liweiyi88/trendshift-backend/database"
)

type RankedTrendingRepository = map[int]TrendingRepository

type TrendingRepositoryRepo struct {
	db database.DB
}

func NewTrendingRepositoryRepo(db database.DB) *TrendingRepositoryRepo {
	return &TrendingRepositoryRepo{
		db: db,
	}
}

// Get all repositories' full name when there is no repository_id set in the table.
func (tr *TrendingRepositoryRepo) FindUnlinkedRepositories(ctx context.Context) ([]string, error) {
	query := "select `full_name` from `trending_repositories` where `repository_id` is null group by `full_name`"

	rows, err := tr.db.QueryContext(ctx, query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	unlinkedRepos := make([]string, 0)

	for rows.Next() {
		var ur struct {
			full_name string
		}

		if err := rows.Scan(&ur.full_name); err != nil {
			return unlinkedRepos, err
		}

		unlinkedRepos = append(unlinkedRepos, ur.full_name)
	}

	if err = rows.Err(); err != nil {
		return unlinkedRepos, err
	}

	return unlinkedRepos, nil
}

func (tr *TrendingRepositoryRepo) FindRankedTrendingRepoByDate(ctx context.Context, date time.Time, language string) (RankedTrendingRepository, error) {
	lang := strings.TrimSpace(language)

	query := "SELECT * FROM trending_repositories WHERE trend_date = ? AND language is null"
	args := []any{date.Format("2006-01-02")}

	if lang != "" {
		query = "SELECT * FROM trending_repositories WHERE trend_date = ? AND language = ?"
		args = append(args, lang)
	}

	rows, err := tr.db.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	rankedTrendingRepositories := make(map[int]TrendingRepository, 0)

	for rows.Next() {
		var tr TrendingRepository

		if err := rows.Scan(&tr.Id, &tr.RepoFullName, &tr.Language, &tr.Rank, &tr.ScrapedAt, &tr.TrendDate, &tr.RepositoryId); err != nil {
			return rankedTrendingRepositories, err
		}

		rankedTrendingRepositories[tr.Rank] = tr
	}

	if err = rows.Err(); err != nil {
		return rankedTrendingRepositories, err
	}

	return rankedTrendingRepositories, nil
}

func (tr *TrendingRepositoryRepo) Save(ctx context.Context, trendingRepository TrendingRepository) error {
	query := "INSERT INTO `trending_repositories` (`full_name`, `language`, `rank`, `scraped_at`, `trend_date`) VALUES (?, ?, ?, ?, ?)"

	scrapeAt := time.Now()

	if !trendingRepository.ScrapedAt.IsZero() {
		scrapeAt = trendingRepository.ScrapedAt
	}

	result, err := tr.db.ExecContext(ctx, query, trendingRepository.RepoFullName, trendingRepository.Language, trendingRepository.Rank, scrapeAt.Format("2006-01-02 15:04:05"), trendingRepository.TrendDate.Format("2006-01-02"))

	if err != nil {
		return fmt.Errorf("failed to exec insert trending_repositories query to db language: %v, full name: %s, error: %v", trendingRepository.Language, trendingRepository.RepoFullName, err)
	}

	_, err = result.RowsAffected()

	if err != nil {
		return fmt.Errorf("trending_repositories insert rows affected returns error: %v", err)
	}

	return nil
}

// Update the trending repository as well as unlink the relationship between tranding repository and repository as the repository might be changed in the same rank.
// Then we leave the linking service to re-link the repository properly.
func (tr *TrendingRepositoryRepo) Update(ctx context.Context, trendingRepository TrendingRepository) error {
	query := "UPDATE `trending_repositories` SET `full_name` = ?, `rank` = ?, `language` = ?, `scraped_at` = ?, `trend_date` = ?, `repository_id` = ? WHERE `id` = ?"

	result, err := tr.db.ExecContext(ctx, query, trendingRepository.RepoFullName, trendingRepository.Rank, trendingRepository.Language, trendingRepository.ScrapedAt.Format("2006-01-02 15:04:05"), trendingRepository.TrendDate.Format("2006-01-02"), nil, trendingRepository.Id)

	if err != nil {
		return fmt.Errorf("failed to run trending_repositories update query, trending repository id: %d, error: %v", trendingRepository.Id, err)
	}

	_, err = result.RowsAffected()

	if err != nil {
		return fmt.Errorf("trending_repositories update rows affected returns error: %v", err)
	}

	return nil
}

// Save the relation between trending repositories and repositories.
func (tr *TrendingRepositoryRepo) LinkRepository(ctx context.Context, repository GhRepository) error {
	query := "UPDATE `trending_repositories` SET repository_id =? WHERE full_name = ?"

	result, err := tr.db.ExecContext(ctx, query, repository.Id, repository.FullName)

	if err != nil {
		return fmt.Errorf("failed to run link repository update query, repository: %s, error: %v", repository.FullName, err)
	}

	_, err = result.RowsAffected()

	if err != nil {
		return fmt.Errorf("link repository rows affected returns error: %v", err)
	}

	return nil
}
