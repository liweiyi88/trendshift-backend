package global

import (
	"github.com/liweiyi88/gti/database"
	"github.com/liweiyi88/gti/model"
	"github.com/liweiyi88/gti/utils/dbutils"
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
	qb := dbutils.NewQueryBuilder()

	return &Repositories{
		TrendingRepositoryRepo: model.NewTrendingRepositoryRepo(db),
		TrendingDeveloperRepo:  model.NewTrendingDeveloperRepo(db),
		DeveloperRepo:          *model.NewDeveloperRepo(db, qb),
		GhRepositoryRepo:       model.NewGhRepositoryRepo(db, qb),
		TagRepo:                model.NewTagRepo(db),
		UserRepo:               model.NewUserRepo(db),
		StatsRepo:              model.NewStatsRepo(db),
	}
}
