package trend

import (
	"context"
	"time"

	"github.com/liweiyi88/gti/internal/database"
)

type TrendRepo struct {
	db database.DB
}

func NewTrendRepo(db database.DB) *TrendRepo {
	return &TrendRepo{
		db: db,
	}
}

func (tr *TrendRepo) FindTrendsByDate(ctx context.Context, date time.Time) ([]Trend, error) {
	rows, err := tr.db.QueryContext(ctx, "SELECT * FROM album WHERE trend_date = ?", date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var trends []Trend

	for rows.Next() {
		var trend Trend

		if err := rows.Scan(&trend.Id, &trend.RepoFullName, &trend.Language, &trend.ScrapedAt, &trend.TrendDate); err != nil {
			return trends, err
		}

		trends = append(trends, trend)
	}

	if err = rows.Err(); err != nil {
		return trends, err
	}

	return trends, nil
}
