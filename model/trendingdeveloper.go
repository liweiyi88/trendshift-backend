package model

import (
	"time"

	"github.com/liweiyi88/trendshift-backend/utils/dbutils"
)

type TrendingDeveloper struct {
	Id          int
	Username    string
	Language    dbutils.NullString
	Rank        int
	ScrapedAt   time.Time
	TrendDate   time.Time
	DeveloperId dbutils.NullInt64
}
