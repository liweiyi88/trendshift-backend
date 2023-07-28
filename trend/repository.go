package trend

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
}

func NewTrendingRepository() *TrendingRepository {
	return &TrendingRepository{}
}
