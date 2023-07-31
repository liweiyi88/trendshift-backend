package github

import (
	"context"
	"fmt"

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

func (gr *GhRepositoryRepo) Save(ctx context.Context, ghRepo GhRepository) error {
	query := "INSERT INTO `gh_repositories` (`full_name`, `ghr_id`, stars, forks, `language`, `owner`, `owner_avatar_url`, `created_at`, `updated_at`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"

	result, err := gr.db.ExecContext(ctx, query,
		ghRepo.FullName,
		ghRepo.GhrId,
		ghRepo.Stars,
		ghRepo.Forks,
		ghRepo.Language,
		ghRepo.Owner.Name,
		ghRepo.Owner.AvatarUrl,
		ghRepo.CreatedAt.Format("2006-01-02 15:04:05"),
		ghRepo.UpdatedAt.Format("2006-01-02 15:04:05"),
	)

	if err != nil {
		return fmt.Errorf("failed to exec insert gh_repositories query to db, error: %v", err)
	}

	_, err = result.RowsAffected()

	if err != nil {
		return fmt.Errorf("gh_repositories insert rows affected returns error: %v", err)
	}

	return nil
}

func (gr *GhRepositoryRepo) Update(ctx context.Context, ghRepo GhRepository) error {
	query := "UPDATE `gh_repositories` SET full_name = ?, ghr_id = ?, stars = ?, forks = ?, language = ?, owner = ?, owner_avatar_url = ?, updated_at = ? WHERE id = ?"

	result, err := gr.db.ExecContext(ctx, query, ghRepo.FullName, ghRepo.GhrId, ghRepo.Stars, ghRepo.Forks, ghRepo.Language, ghRepo.Owner.Name, ghRepo.Owner.AvatarUrl, ghRepo.UpdatedAt.Format("2006-01-02 15:04:05"), ghRepo.Id)

	if err != nil {
		return fmt.Errorf("failed to run gh_repositories update query, gh repo id: %d, error: %v", ghRepo.Id, err)
	}

	_, err = result.RowsAffected()

	if err != nil {
		return fmt.Errorf("gh_repositories update rows affected returns error: %v", err)
	}

	return nil
}
