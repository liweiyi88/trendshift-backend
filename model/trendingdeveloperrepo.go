package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/liweiyi88/trendshift-backend/database"
)

type RankedTrendingDevelopers = map[int]TrendingDeveloper

type TrendingDeveloperRepo struct {
	db database.DB
}

func NewTrendingDeveloperRepo(db database.DB) *TrendingDeveloperRepo {
	return &TrendingDeveloperRepo{
		db,
	}
}

func (tdr *TrendingDeveloperRepo) LinkDeveloper(ctx context.Context, developer Developer) error {
	query := "UPDATE `trending_developers` SET developer_id =? WHERE username = ?"

	result, err := tdr.db.ExecContext(ctx, query, developer.Id, developer.Username)

	if err != nil {
		return fmt.Errorf("failed to run link developer update query, developer: %s, error: %v", developer.Username, err)
	}

	_, err = result.RowsAffected()

	if err != nil {
		return fmt.Errorf("link developer rows affected returns error: %v", err)
	}

	return nil
}

func (tdr *TrendingDeveloperRepo) FindUnlinkedDevelopers(ctx context.Context) ([]string, error) {
	query := "select `username` from `trending_developers` where `developer_id` is null group by `username`"

	rows, err := tdr.db.QueryContext(ctx, query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	unlinkedDevelopers := make([]string, 0)

	for rows.Next() {
		var ur struct {
			username string
		}

		if err := rows.Scan(&ur.username); err != nil {
			return unlinkedDevelopers, err
		}

		unlinkedDevelopers = append(unlinkedDevelopers, ur.username)
	}

	if err = rows.Err(); err != nil {
		return unlinkedDevelopers, err
	}

	return unlinkedDevelopers, nil

}

func (tdr *TrendingDeveloperRepo) FindRankedTrendingDevelopersByDate(ctx context.Context, date time.Time, language string) (RankedTrendingDevelopers, error) {
	lang := strings.TrimSpace(language)

	query := "SELECT * FROM trending_developers WHERE trend_date = ? AND language is null"
	args := []any{date.Format("2006-01-02")}

	if lang != "" {
		query = "SELECT * FROM trending_developers WHERE trend_date = ? AND language = ?"
		args = append(args, lang)
	}

	rows, err := tdr.db.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	rankedTrendingDevelopers := make(map[int]TrendingDeveloper, 0)

	for rows.Next() {
		var td TrendingDeveloper

		if err := rows.Scan(&td.Id, &td.Username, &td.Language, &td.Rank, &td.ScrapedAt, &td.TrendDate, &td.DeveloperId); err != nil {
			return rankedTrendingDevelopers, err
		}

		rankedTrendingDevelopers[td.Rank] = td
	}

	if err = rows.Err(); err != nil {
		return rankedTrendingDevelopers, err
	}

	return rankedTrendingDevelopers, nil
}

func (tdr *TrendingDeveloperRepo) Save(ctx context.Context, trendingDeveloper TrendingDeveloper) error {
	query := "INSERT INTO `trending_developers` (`username`, `language`, `rank`, `scraped_at`, `trend_date`) VALUES (?, ?, ?, ?, ?)"

	scrapeAt := time.Now()

	if !trendingDeveloper.ScrapedAt.IsZero() {
		scrapeAt = trendingDeveloper.ScrapedAt
	}

	result, err := tdr.db.ExecContext(
		ctx,
		query,
		trendingDeveloper.Username,
		trendingDeveloper.Language,
		trendingDeveloper.Rank,
		scrapeAt.Format("2006-01-02 15:04:05"),
		trendingDeveloper.TrendDate.Format("2006-01-02"),
	)

	if err != nil {
		return fmt.Errorf("failed to exec insert trending_developers query to db language: %v, username: %s, error: %v", trendingDeveloper.Language, trendingDeveloper.Username, err)
	}

	_, err = result.RowsAffected()

	if err != nil {
		return fmt.Errorf("trending_developers insert rows affected returns error: %v", err)
	}

	return nil
}

func (tdr *TrendingDeveloperRepo) Update(ctx context.Context, trendingDeveloper TrendingDeveloper) error {
	query := "UPDATE `trending_developers` SET username = ?, rank = ?, language = ?, scraped_at = ?, trend_date = ?, developer_id = ? WHERE id = ?"

	result, err := tdr.db.ExecContext(
		ctx,
		query,
		trendingDeveloper.Username,
		trendingDeveloper.Rank,
		trendingDeveloper.Language,
		trendingDeveloper.ScrapedAt.Format("2006-01-02 15:04:05"),
		trendingDeveloper.TrendDate.Format("2006-01-02"),
		nil,
		trendingDeveloper.Id,
	)

	if err != nil {
		return fmt.Errorf("failed to run trending_developers update query, trending developer id: %d, error: %v", trendingDeveloper.Id, err)
	}

	_, err = result.RowsAffected()

	if err != nil {
		return fmt.Errorf("trending_developers update rows affected returns error: %v", err)
	}

	return nil
}
