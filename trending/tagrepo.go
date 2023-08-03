package trending

import (
	"context"
	"fmt"
	"strings"

	"github.com/liweiyi88/gti/database"
)

type TagRepo struct {
	db database.DB
}

func NewTagRepositoryRepo(db database.DB) *TagRepo {
	return &TagRepo{
		db: db,
	}
}

func (tr *TagRepo) Find(ctx context.Context, name string) ([]Tag, error) {
	var query string
	args := []any{}

	if strings.TrimSpace(name) == "" {
		query = "SELECT * FROM tags"
	} else {
		query = "SELECT * FROM tags WHERE name like ? LIMIT 10"
		args = append(args, "%"+name+"%")
	}

	rows, err := tr.db.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	tags := make([]Tag, 0)

	for rows.Next() {
		var tag Tag

		if err := rows.Scan(&tag.Id, &tag.Name); err != nil {
			return tags, err
		}

		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return tags, err
	}

	return tags, nil
}

func (tr *TagRepo) FindByName(ctx context.Context, name string) (Tag, error) {
	query := "SELECT * FROM tags WHERE LOWER(name) = ?"

	var tag Tag

	row := tr.db.QueryRowContext(ctx, query, name)

	if err := row.Scan(&tag.Id, &tag.Name); err != nil {
		return tag, err
	}

	return tag, nil
}

func (tr *TagRepo) Save(ctx context.Context, tag Tag) (int, error) {
	query := "INSERT INTO `tags` (`name`) VALUES (?)"

	var lastInsertId int64

	result, err := tr.db.ExecContext(ctx, query, tag.Name)

	if err != nil {
		return int(lastInsertId), fmt.Errorf("failed to exec insert tags query to db, error: %v", err)
	}

	lastInsertId, err = result.LastInsertId()
	if err != nil {
		return int(lastInsertId), fmt.Errorf("failed to get tags last insert id after insert, error: %v", err)
	}

	_, err = result.RowsAffected()

	if err != nil {
		return int(lastInsertId), fmt.Errorf("tag insert rows affected returns error: %v", err)
	}

	return int(lastInsertId), nil
}
