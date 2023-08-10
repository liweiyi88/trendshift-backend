package model

import (
	"context"

	"github.com/liweiyi88/gti/database"
)

type UserRepo struct {
	db database.DB
}

func NewUserRepo(db database.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (ur *UserRepo) FindByNamePassword(ctx context.Context, username, hashedPassword string) (User, error) {
	query := "SELECT * FROM users WHERE username = ? AND password = ?"

	var user User

	row := ur.db.QueryRowContext(ctx, query, username, hashedPassword)

	if err := row.Scan(&user.Id, &user.Username, &user.Password, &user.Role); err != nil {
		return user, err
	}

	return user, nil
}
