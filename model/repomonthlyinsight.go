package model

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/liweiyi88/trendshift-backend/database"
	"github.com/liweiyi88/trendshift-backend/utils/dbutils"
)

type RepositoryMonthlyInsight struct {
	Id             int               `json:"id"`
	Year           int               `json:"year"`
	Month          int               `json:"month"`
	Stars          dbutils.NullInt64 `json:"stars"`
	Forks          dbutils.NullInt64 `json:"forks"`
	MergedPrs      dbutils.NullInt64 `json:"merged_prs"`
	Issues         dbutils.NullInt64 `json:"issues"`
	ClosedIssues   dbutils.NullInt64 `json:"closed_issues"`
	CompletedAt    dbutils.NullTime  `json:"completed_at"`
	LastIngestedAt dbutils.NullTime  `json:"last_ingested_at"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	RepositoryId   int               `json:"repository_id"`
}

type RepositoryMonthlyInsightWithName struct {
	RepositoryMonthlyInsight
	RepositoryName string
}

type RepositoryMonthlyInsightRepo struct {
	db database.DB
}

func NewRepositoryMonthlyInsightRepo(db database.DB) *RepositoryMonthlyInsightRepo {
	return &RepositoryMonthlyInsightRepo{
		db: db,
	}
}

func (rr *RepositoryMonthlyInsightRepo) CreateMonthlyInsightsIfNotExist(ctx context.Context, month, year int) (int64, error) {
	query := `
    INSERT INTO repository_monthly_insights (repository_id, month, year, created_at, updated_at)
	SELECT r.id, ?, ?, ?, ?
    FROM repositories r
    LEFT JOIN repository_monthly_insights mi
      ON mi.repository_id = r.id AND mi.month = ? AND mi.year = ?
    WHERE mi.id IS NULL AND r.skipped = false;
`

	createdAt, updatedAt := time.Now(), time.Now()

	result, err := rr.db.ExecContext(ctx, query,
		month,
		year,
		createdAt.Format(time.DateTime),
		updatedAt.Format(time.DateTime),
		month,
		year,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to exec insert repository_monthly_insights query to db, error: %v", err)
	}

	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return lastInsertId, fmt.Errorf("failed to get repository_monthly_insights last insert id after insert, error: %v", err)
	}

	_, err = result.RowsAffected()

	if err != nil {
		return 0, fmt.Errorf("repository_monthly_insights insert rows affected returns error: %v", err)
	}

	return lastInsertId, nil
}

func (rr *RepositoryMonthlyInsightRepo) Update(ctx context.Context, data RepositoryMonthlyInsightWithName) error {
	query := "UPDATE `repository_monthly_insights` SET year = ?, month = ?, stars = ?, forks = ?, merged_prs = ?, issues = ?, closed_issues = ?, completed_at = ?, last_ingested_at = ?, updated_at = ? WHERE id = ?"

	updatedAt := time.Now()

	var completedAt, lastIngestedAt any

	if data.CompletedAt.Valid {
		completedAt = data.CompletedAt.Time.Format(time.DateTime)
	} else {
		completedAt = nil
	}

	if data.LastIngestedAt.Valid {
		lastIngestedAt = data.LastIngestedAt.Time.Format(time.DateTime)
	} else {
		lastIngestedAt = nil
	}

	result, err := rr.db.ExecContext(
		ctx, query,
		data.Year,
		data.Month,
		data.Stars,
		data.Forks,
		data.MergedPrs,
		data.Issues,
		data.ClosedIssues,
		completedAt,
		lastIngestedAt,
		updatedAt.Format(time.DateTime),
		data.Id)

	if err != nil {
		return fmt.Errorf("failed to run repository_monthly_insights update query, id: %d, error: %v", data.Id, err)
	}

	n, err := result.RowsAffected()

	if err != nil {
		return fmt.Errorf("repository_monthly_insights update rows affected returns error: %v", err)
	}

	if n != 1 {
		return fmt.Errorf("unexpected number of rows affected after update: %d", n)
	}

	return nil
}

func (rr *RepositoryMonthlyInsightRepo) FindIncompletedLastIngestedBefore(ctx context.Context, before time.Time, limit int) ([]RepositoryMonthlyInsightWithName, error) {
	query := "select ri.*, repositories.full_name from repository_monthly_insights as ri JOIN repositories ON ri.repository_id = repositories.id where ri.completed_at is null AND repositories.skipped = false AND (ri.last_ingested_at is null OR ri.last_ingested_at < ?) order by ri.month ASC, ri.last_ingested_at ASC limit ?"
	args := []any{before.Format(time.DateTime), limit}

	rows, err := rr.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find incompleted repository monthly insight before %s, error: %v", before.Format(time.DateTime), err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("failed to close rows", slog.Any("error", err), slog.String("action", "repositoryMonthlyInsightRepo.FindIncompleted"))
		}
	}()

	data := make([]RepositoryMonthlyInsightWithName, 0)

	for rows.Next() {
		var repoInsight RepositoryMonthlyInsightWithName

		if err := rows.Scan(
			&repoInsight.Id,
			&repoInsight.Year,
			&repoInsight.Month,
			&repoInsight.Stars,
			&repoInsight.Forks,
			&repoInsight.MergedPrs,
			&repoInsight.Issues,
			&repoInsight.ClosedIssues,
			&repoInsight.CompletedAt,
			&repoInsight.LastIngestedAt,
			&repoInsight.CreatedAt,
			&repoInsight.UpdatedAt,
			&repoInsight.RepositoryId,
			&repoInsight.RepositoryName); err != nil {
			return nil, fmt.Errorf("failed to scan repository_monthly_insights table, error: %v", err)
		}

		data = append(data, repoInsight)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("repositoryMonthlyInsightRepo.FindIncompleted, rows error: %v", err)
	}

	return data, nil
}
