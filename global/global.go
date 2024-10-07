package global

import (
	"github.com/liweiyi88/trendshift-backend/database"
	"github.com/liweiyi88/trendshift-backend/model"
)

type Repositories struct {
	TrendingRepositoryRepo *model.TrendingRepositoryRepo
	TrendingDeveloperRepo  *model.TrendingDeveloperRepo
	DeveloperRepo          model.DeveloperRepo
	GhRepositoryRepo       *model.GhRepositoryRepo
	TagRepo                *model.TagRepo
	UserRepo               *model.UserRepo
	StatsRepo              *model.StatsRepo
}

func InitRepositories(db database.DB) *Repositories {
	return &Repositories{
		TrendingRepositoryRepo: model.NewTrendingRepositoryRepo(db),
		TrendingDeveloperRepo:  model.NewTrendingDeveloperRepo(db),
		DeveloperRepo:          *model.NewDeveloperRepo(db),
		GhRepositoryRepo:       model.NewGhRepositoryRepo(db),
		TagRepo:                model.NewTagRepo(db),
		UserRepo:               model.NewUserRepo(db),
		StatsRepo:              model.NewStatsRepo(db),
	}
}
