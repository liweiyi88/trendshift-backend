package global

import (
	"github.com/liweiyi88/gti/database"
	"github.com/liweiyi88/gti/model"
)

type Repositories struct {
	TrendingRepositoryRepo *model.TrendingRepositoryRepo
	GhRepositoryRepo       *model.GhRepositoryRepo
	TagRepo                *model.TagRepo
	UserRepo               *model.UserRepo
	StatsRepo              *model.StatsRepo
}

func InitRepositories(db database.DB) *Repositories {
	return &Repositories{
		TrendingRepositoryRepo: model.NewTrendingRepositoryRepo(db),
		GhRepositoryRepo:       model.NewGhRepositoryRepo(db),
		TagRepo:                model.NewTagRepo(db),
		UserRepo:               model.NewUserRepo(db),
		StatsRepo:              model.NewStatsRepo(db),
	}
}
