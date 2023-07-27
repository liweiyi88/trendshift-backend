package trend

import (
	"context"
	"fmt"
	"time"

	"github.com/liweiyi88/gti/internal/database"
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

func (tr *TrendingRepositoryRepo) FindRankedTrendsByDate(ctx context.Context, date time.Time) (RankedTrendingRepository, error) {
	rows, err := tr.db.QueryContext(ctx, "SELECT * FROM album WHERE trend_date = ?", date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	rankedTrends := make(map[int]TrendingRepository, 0)

	for rows.Next() {
		var trend TrendingRepository

		if err := rows.Scan(&trend.Id, &trend.RepoFullName, &trend.Language, &trend.ScrapedAt, &trend.TrendDate); err != nil {
			return rankedTrends, err
		}

		rankedTrends[trend.Rank] = trend
	}

	if err = rows.Err(); err != nil {
		return rankedTrends, err
	}

	return rankedTrends, nil
}

func (tr *TrendingRepositoryRepo) Save(ctx context.Context, trend TrendingRepository) error {
	query := "INSERT INTO `trending_repositories` (`full_name`, `language`, `rank`, `scraped_at`, `trend_date`) VALUES (?, ?, ?, ?, ?)"

	_, err := tr.db.ExecContext(ctx, query, trend.RepoFullName, trend.Language, trend.Rank, trend.ScrapedAt.Format("2006-01-02 15:04:05"), trend.TrendDate.Format("2006-01-02"))

	if err != nil {
		return fmt.Errorf("failed to save trend to db, error: %v", err)
	}

	return nil
}

func (tr *TrendingRepositoryRepo) Update(ctx context.Context, trend TrendingRepository) error {
	query := "UPDATE `trending_repositories` SET full_name = ?, rank = ?, language = ?, scraped_at = ?, trend_date = ? WHERE id = ?"

	_, err := tr.db.ExecContext(ctx, query, trend.RepoFullName, trend.Rank, trend.Language, trend.ScrapedAt.Format("2006-01-02 15:04:05"), trend.TrendDate.Format("2006-01-02"), trend.Id)

	if err != nil {
		return fmt.Errorf("failed to update trend to db, trend id: %d, error: %v", trend.Id, err)
	}

	return nil
}
