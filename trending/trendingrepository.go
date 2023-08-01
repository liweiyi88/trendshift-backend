package trending

import (
	"database/sql"
	"time"
)

type TrendingRepository struct {
	Id           int
	RepoFullName string
	Language     sql.NullString
	Rank         int
	ScrapedAt    time.Time
	TrendDate    time.Time
	RepositoryId sql.NullInt64
}
