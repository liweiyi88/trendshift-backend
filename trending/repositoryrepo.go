package trending

import (
	"context"
	"fmt"
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

func (gr *GhRepositoryRepo) FindRepositoriesByNames(ctx context.Context, names []string) ([]GhRepository, error) {
	ghRepos := make([]GhRepository, 0)

	if len(names) == 0 {
		return ghRepos, nil
	}

	query := fmt.Sprintf("SELECT * FROM repositories WHERE full_name in (\"%s\")", strings.Join(names, "\",\""))

	rows, err := gr.db.QueryContext(ctx, query)

	if err != nil {
		return nil, err
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

func (gr *GhRepositoryRepo) UpdateWithTags(ctx context.Context, ghRepo GhRepository, tags []Tag) error {
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
