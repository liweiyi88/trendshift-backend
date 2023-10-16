package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/liweiyi88/trendshift-backend/database"
	"github.com/liweiyi88/trendshift-backend/utils/dbutils"
)

type DeveloperRepo struct {
	db database.DB
	qb *dbutils.QueryBuilder
}

func NewDeveloperRepo(db database.DB, qb *dbutils.QueryBuilder) *DeveloperRepo {
	return &DeveloperRepo{db, qb}
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
