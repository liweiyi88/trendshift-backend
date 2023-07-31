package repository

import (
	"github.com/liweiyi88/gti/database"
	"github.com/liweiyi88/gti/trending"
)

type Repositories struct {
	TrendingRepositoryRepo *trend.TrendingRepositoryRepo
}

func InitRepositories(db database.DB) *Repositories {
	trendingRepositoryRepo := trend.NewTrendingRepositoryRepo(db)

	return &Repositories{
		TrendingRepositoryRepo: trendingRepositoryRepo,
	}
}
