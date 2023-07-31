package repository

import (
	"github.com/liweiyi88/gti/database"
	"github.com/liweiyi88/gti/github"
	trend "github.com/liweiyi88/gti/trending"
)

type Repositories struct {
	TrendingRepositoryRepo *trend.TrendingRepositoryRepo
	GhRepositoryRepo       *github.GhRepositoryRepo
}

func InitRepositories(db database.DB) *Repositories {
	return &Repositories{
		TrendingRepositoryRepo: trend.NewTrendingRepositoryRepo(db),
		GhRepositoryRepo:       github.NewGhRepositoryRepo(db),
	}
}
