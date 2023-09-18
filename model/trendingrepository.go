package model

import (
	"time"

	"github.com/liweiyi88/gti/dbutils"
)

type TrendingRepository struct {
	Id           int
	RepoFullName string
	Language     dbutils.NullString
	Rank         int
	ScrapedAt    time.Time
	TrendDate    time.Time
	RepositoryId dbutils.NullInt64
}
