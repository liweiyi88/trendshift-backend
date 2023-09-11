package model

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/liweiyi88/gti/database"
)

type GhRepositoryRepo struct {
	db database.DB
}

func NewGhRepositoryRepo(db database.DB) *GhRepositoryRepo {
	return &GhRepositoryRepo{
		db: db,
	}
}

func (gr *GhRepositoryRepo) FindById(ctx context.Context, id int) (GhRepository, error) {
	query := "SELECT * FROM repositories WHERE id = ?"

	var ghr GhRepository

	row := gr.db.QueryRowContext(ctx, query, id)

	if err := row.Scan(&ghr.Id, &ghr.GhrId, &ghr.Stars, &ghr.Forks, &ghr.FullName, &ghr.Language, &ghr.Owner.Name, &ghr.Owner.AvatarUrl, &ghr.CreatedAt, &ghr.UpdatedAt); err != nil {
		return ghr, err
	}

	return ghr, nil
}

func (gr *GhRepositoryRepo) FindAllWithTags(ctx context.Context, filter string) ([]GhRepository, error) {
	var query string
	var args []any

	if filter == "today" {
		query = "select repositories.*, tags.id as tag_id, tags.`name` as tag_name  from repositories left join repositories_tags ON repositories.id = repositories_tags.repository_id left join tags on repositories_tags.tag_id = tags.id join trending_repositories on repositories.id = trending_repositories.repository_id where trend_date = ? group by repositories.full_name, tags.id"
		args = append(args, time.Now().Format("2006-01-02"))
	} else {
		query = "select repositories.*, tags.id as tag_id, tags.`name` as tag_name from repositories left join repositories_tags ON repositories.id = repositories_tags.repository_id left join tags on repositories_tags.tag_id = tags.id"
	}

	rows, err := gr.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	repoMap := make(map[int]*GhRepository, 0)

	for rows.Next() {
		var ghr GhRepository
		var tagId sql.NullInt64
		var tagName sql.NullString

		if err := rows.Scan(&ghr.Id, &ghr.GhrId, &ghr.Stars, &ghr.Forks, &ghr.FullName, &ghr.Language, &ghr.Owner.Name, &ghr.Owner.AvatarUrl, &ghr.CreatedAt, &ghr.UpdatedAt, &tagId, &tagName); err != nil {
			return nil, err
		}

		_, ok := repoMap[ghr.Id]

		if !ok {
			ghr.Tags = make([]Tag, 0)
			repoMap[ghr.Id] = &ghr
		}

		if tagId.Valid && tagName.Valid {
			tag := Tag{
				Id:   int(tagId.Int64),
				Name: tagName.String,
			}

			repoMap[ghr.Id].Tags = append(repoMap[ghr.Id].Tags, tag)
		}
	}

	ghRepos := make([]GhRepository, 0, len(repoMap))

	keys := make([]int, 0, len(repoMap))
	for k := range repoMap {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	for _, k := range keys {
		ghRepos = append(ghRepos, *repoMap[k])
	}

	return ghRepos, nil
}

func (gr *GhRepositoryRepo) FindRepositoriesByNames(ctx context.Context, names []string) ([]GhRepository, error) {
	ghRepos := make([]GhRepository, 0)

	if len(names) == 0 {
		return ghRepos, nil
	}

	query := fmt.Sprintf("SELECT * FROM repositories WHERE full_name IN ('%s')", strings.Join(names, "','"))

	rows, err := gr.db.QueryContext(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("failed to run the select query: %s, error: %v", query, err)
	}

	defer rows.Close()

	for rows.Next() {
		var ghr GhRepository

		if err := rows.Scan(&ghr.Id, &ghr.GhrId, &ghr.Stars, &ghr.Forks, &ghr.FullName, &ghr.Language, &ghr.Owner.Name, &ghr.Owner.AvatarUrl, &ghr.CreatedAt, &ghr.UpdatedAt); err != nil {
			return ghRepos, err
		}

		ghRepos = append(ghRepos, ghr)
	}

	if err = rows.Err(); err != nil {
		return ghRepos, err
	}

	return ghRepos, nil
}

func (gr *GhRepositoryRepo) Save(ctx context.Context, ghRepo GhRepository) (int64, error) {
	query := "INSERT INTO `repositories` (`full_name`, `ghr_id`, stars, forks, `language`, `owner`, `owner_avatar_url`, `created_at`, `updated_at`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"

	var lastInsertId int64

	createdAt, updatedAt := time.Now(), time.Now()

	if !ghRepo.CreatedAt.IsZero() {
		createdAt = ghRepo.CreatedAt
	}

	if !ghRepo.UpdatedAt.IsZero() {
		updatedAt = ghRepo.UpdatedAt
	}

	result, err := gr.db.ExecContext(ctx, query,
		ghRepo.FullName,
		ghRepo.GhrId,
		ghRepo.Stars,
		ghRepo.Forks,
		ghRepo.Language,
		ghRepo.Owner.Name,
		ghRepo.Owner.AvatarUrl,
		createdAt.Format("2006-01-02 15:04:05"),
		updatedAt.Format("2006-01-02 15:04:05"),
	)

	if err != nil {
		return lastInsertId, fmt.Errorf("failed to exec insert repositories query to db, error: %v", err)
	}

	lastInsertId, err = result.LastInsertId()
	if err != nil {
		return lastInsertId, fmt.Errorf("failed to get repositories last insert id after insert, error: %v", err)
	}

	_, err = result.RowsAffected()

	if err != nil {
		return lastInsertId, fmt.Errorf("repositories insert rows affected returns error: %v", err)
	}

	return lastInsertId, nil
}

func (gr *GhRepositoryRepo) Update(ctx context.Context, ghRepo GhRepository) error {
	query := "UPDATE `repositories` SET full_name = ?, ghr_id = ?, stars = ?, forks = ?, language = ?, owner = ?, owner_avatar_url = ?, updated_at = ? WHERE id = ?"

	updatedAt := time.Now()

	if !ghRepo.UpdatedAt.IsZero() {
		updatedAt = ghRepo.UpdatedAt
	}

	result, err := gr.db.ExecContext(ctx, query, ghRepo.FullName, ghRepo.GhrId, ghRepo.Stars, ghRepo.Forks, ghRepo.Language, ghRepo.Owner.Name, ghRepo.Owner.AvatarUrl, updatedAt.Format("2006-01-02 15:04:05"), ghRepo.Id)

	if err != nil {
		return fmt.Errorf("failed to run repositories update query, gh repo id: %d, error: %v", ghRepo.Id, err)
	}

	_, err = result.RowsAffected()

	if err != nil {
		return fmt.Errorf("repositories update rows affected returns error: %v", err)
	}

	return nil
}

func (gr *GhRepositoryRepo) SaveTags(ctx context.Context, ghRepo GhRepository, tags []Tag) error {
	tx, err := gr.db.BeginTx(ctx, nil)

	if err != nil {
		return fmt.Errorf("failed to begin repository update tags transaction: %v", err)
	}

	defer tx.Rollback()
	query := "DELETE FROM `repositories_tags` WHERE repository_id = ?"

	_, err = tx.ExecContext(ctx, query, ghRepo.Id)

	if err != nil {
		return fmt.Errorf("failed to run delete repositories_tags query, repository id: %d, error: %v", ghRepo.Id, err)
	}

	for _, tag := range tags {
		query := "INSERT INTO `repositories_tags` (`repository_id`, `tag_id`) VALUES (?, ?)"

		result, err := tx.ExecContext(ctx, query, ghRepo.Id, tag.Id)

		if err != nil {
			return fmt.Errorf("failed to run insert repositories_tags query, repository id: %d, tag id: %d error: %v", ghRepo.Id, tag.Id, err)
		}

		_, err = result.RowsAffected()

		if err != nil {
			return fmt.Errorf("repositories_tags insert rows affected returns error: %v", err)
		}
	}

	return tx.Commit()
}
