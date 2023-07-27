package trend

import "time"

type Trend struct {
	Id           int
	RepoFullName string
	Language     string
	ScrapedAt    time.Time
	TrendDate   time.Time
}
