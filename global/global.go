package global

import (
	"github.com/liweiyi88/gti/database"
	"github.com/liweiyi88/gti/trending"
)

type Repositories struct {
	TrendingRepositoryRepo *trending.TrendingRepositoryRepo
	GhRepositoryRepo       *trending.GhRepositoryRepo
}

func InitRepositories(db database.DB) *Repositories {
	return &Repositories{
		TrendingRepositoryRepo: trending.NewTrendingRepositoryRepo(db),
		GhRepositoryRepo:       trending.NewGhRepositoryRepo(db),
	}
}
