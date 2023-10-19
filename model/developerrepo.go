package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/liweiyi88/trendshift-backend/database"
	"github.com/liweiyi88/trendshift-backend/model/opt"
	"github.com/liweiyi88/trendshift-backend/utils/dbutils"
)

type TrendingDeveloperResponse struct {
	Developer
	BestRanking   int `json:"best_ranking"`   // non db column field
	FeaturedCount int `json:"featured_count"` // non db column field
}

type DeveloperRepo struct {
	db database.DB
	qb *dbutils.QueryBuilder
}

func NewDeveloperRepo(db database.DB, qb *dbutils.QueryBuilder) *DeveloperRepo {
	return &DeveloperRepo{db, qb}
}

func (dr *DeveloperRepo) FindById(ctx context.Context, id int) (Developer, error) {
	qb := dr.qb
	qb.Query("select developers.*, trending_developers.`trend_date`, trending_developers.`rank`, trending_developers.`language` as `trending_language` from developers join trending_developers on developers.id = trending_developers.developer_id")
	qb.Where("developers.id = ?", id)
	query, args := qb.GetQuery()

	var developer Developer

	rows, err := dr.db.QueryContext(ctx, query, args...)

	if err != nil {
		return developer, fmt.Errorf("failed to find developer by id: %v", err)
	}

	defer rows.Close()

	collectionMap := dbutils.NewCollectionMap[int, *Developer]()

	for rows.Next() {
		var trending Trending

		if err := rows.Scan(
			&developer.Id,
			&developer.GhId,
			&developer.Username,
			&developer.AvatarUrl,
			&developer.Name,
			&developer.Company,
			&developer.Blog,
			&developer.Location,
			&developer.Email,
			&developer.Bio,
			&developer.TwitterUsername,
			&developer.PublicRepos,
			&developer.PublicGists,
			&developer.Followers,
			&developer.Following,
			&developer.CreatedAt,
			&developer.UpdatedAt,
			&trending.TrendDate,
			&trending.Rank,
			&trending.TrendingLanguage,
			); err != nil {
			return developer, err
		}

		if !collectionMap.Has(developer.Id) {
			developer.Trendings = append(developer.Trendings, trending)
			collectionMap.Set(developer.Id, &developer)
		} else {
			developer := collectionMap.Get(developer.Id)
			developer.Trendings = append(developer.Trendings, trending)
		}
	}

	if err = rows.Err(); err != nil {
		return developer, err
	}

	return developer, nil
}

func (dr *DeveloperRepo) FindTrendingDevelopers(ctx context.Context, opts ...any) ([]TrendingDeveloperResponse, error) {
	query := "select developers.*, count(*) as count, min(trending_developers.`rank`) as best_ranking from developers join trending_developers on developers.id = trending_developers.developer_id"

	qb := dr.qb
	qb.Query(query)

	qb.OrderBy("count", "DESC")
	qb.OrderBy("best_ranking", "ASC")
	qb.OrderBy("developers.id", "ASC")

	options := opt.ExtractOptions(opts...)
	lang, dateRange, limit := options.Language, options.DateRange, options.Limit

	if lang != "" {
		qb.Where("`trending_developers`.`language` = ?", lang)
	} else {
		qb.Where("`trending_developers`.`language` is null", nil)
	}

	if dateRange > 0 {
		since := time.Now().AddDate(0, 0, -dateRange)
		qb.Where("`trending_developers`.`trend_date` > ?", since.Format("2006-01-02"))
	}

	if limit > 0 {
		qb.Limit(limit)
	}

	qb.GroupBy("developers.id")

	q, args := qb.GetQuery()

	rows, err := dr.db.QueryContext(ctx, q, args...)

	if err != nil {
		return nil, fmt.Errorf("failed to query trending developers: %v", err)
	}

	defer rows.Close()

	var developers []TrendingDeveloperResponse

	for rows.Next() {
		var dev TrendingDeveloperResponse

		if err := rows.Scan(
			&dev.Id,
			&dev.GhId,
			&dev.Username,
			&dev.AvatarUrl,
			&dev.Name,
			&dev.Company,
			&dev.Blog,
			&dev.Location,
			&dev.Email,
			&dev.Bio,
			&dev.TwitterUsername,
			&dev.PublicRepos,
			&dev.PublicGists,
			&dev.Followers,
			&dev.Following,
			&dev.CreatedAt,
			&dev.UpdatedAt,
			&dev.FeaturedCount,
			&dev.BestRanking,
		); err != nil {
			return nil, err

		}

		developers = append(developers, dev)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return developers, nil
}

func (dr *DeveloperRepo) Save(ctx context.Context, developer Developer) (int64, error) {
	query := "INSERT INTO `developers` (`gh_id`, `username`, `avatar_url`, `name`, `company`, `blog`, `location`, `email`, `bio`, `twitter_username`, `public_repos`, `public_gists`, `followers`, `following`, `created_at`, `updated_at`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

	var lastInsertId int64

	createdAt, updatedAt := time.Now(), time.Now()

	result, err := dr.db.ExecContext(ctx, query,
		developer.GhId,
		developer.Username,
		developer.AvatarUrl,
		developer.Name,
		developer.Company,
		developer.Blog,
		developer.Location,
		developer.Email,
		developer.Bio,
		developer.TwitterUsername,
		developer.PublicRepos,
		developer.PublicGists,
		developer.Followers,
		developer.Following,
		createdAt.Format("2006-01-02 15:04:05"),
		updatedAt.Format("2006-01-02 15:04:05"),
	)

	if err != nil {
		return lastInsertId, fmt.Errorf("failed to exec insert developers query to db, error: %v", err)
	}

	lastInsertId, err = result.LastInsertId()
	if err != nil {
		return lastInsertId, fmt.Errorf("failed to get developers last insert id after insert, error: %v", err)
	}

	_, err = result.RowsAffected()

	if err != nil {
		return lastInsertId, fmt.Errorf("developers insert rows affected returns error: %v", err)
	}

	return lastInsertId, nil
}

func (dr *DeveloperRepo) FindDevelopersByUsernames(ctx context.Context, names []string) ([]Developer, error) {
	developers := make([]Developer, 0)

	if len(names) == 0 {
		return developers, nil
	}

	query := fmt.Sprintf("SELECT * FROM developers WHERE username IN ('%s')", strings.Join(names, "','"))

	rows, err := dr.db.QueryContext(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("failed to run the select query: %s, error: %v", query, err)
	}

	defer rows.Close()

	for rows.Next() {
		var developer Developer

		if err := rows.Scan(
			&developer.Id,
			&developer.GhId,
			&developer.Username,
			&developer.AvatarUrl,
			&developer.Name,
			&developer.Company,
			&developer.Blog,
			&developer.Location,
			&developer.Email,
			&developer.Bio,
			&developer.TwitterUsername,
			&developer.PublicRepos,
			&developer.PublicGists,
			&developer.Followers,
			&developer.Following,
			&developer.CreatedAt,
			&developer.UpdatedAt); err != nil {
			return developers, err
		}

		developers = append(developers, developer)
	}

	if err = rows.Err(); err != nil {
		return developers, err
	}

	return developers, nil
}
