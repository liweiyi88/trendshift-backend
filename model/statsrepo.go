package model

import (
	"context"
	"time"

	"github.com/liweiyi88/gti/database"
)

type DailyStat struct {
	Count     int       `json:"count"`
	Name      string    `json:"name"`
	TrendDate time.Time `json:"trend_date"`
}

type StatsRepo struct {
	db database.DB
}

func NewStatsRepo(db database.DB) *StatsRepo {
	return &StatsRepo{
		db: db,
	}
}

func (sr *StatsRepo) FindDailyStats(ctx context.Context) ([]DailyStat, error) {
	query := "select count(*) as count, tags.`name`, trend_date from trending_repositories JOIN repositories ON trending_repositories.repository_id = repositories.id join repositories_tags on repositories_tags.repository_id = repositories.id join tags on tags.id = repositories_tags.tag_id group by tags.`name`, trend_date order by trend_date ASC"

	rows, err := sr.db.QueryContext(ctx, query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	dailyStats := make([]DailyStat, 0)

	for rows.Next() {
		var dailyStat DailyStat

		if err := rows.Scan(&dailyStat.Count, &dailyStat.Name, &dailyStat.TrendDate); err != nil {
			return nil, err
		}

		dailyStats = append(dailyStats, dailyStat)
	}

	if err = rows.Err(); err != nil {
		return dailyStats, err
	}

	return dailyStats, nil
}
