package model

import (
	"context"
	"fmt"
	"time"

	"github.com/liweiyi88/trendshift-backend/database"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id        int
	Username  string
	Password  string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (user *User) IsPasswordValid(plainPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(plainPassword)) == nil
}

func (user *User) SetPassword(plainPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)

	return nil
}

type UserRepo struct {
	db database.DB
}

func NewUserRepo(db database.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (ur *UserRepo) FindByName(ctx context.Context, username string) (User, error) {
	query := "SELECT * FROM users WHERE username = ?"

	var user User

	row := ur.db.QueryRowContext(ctx, query, username)

	if err := row.Scan(&user.Id, &user.Username, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return user, err
	}

	return user, nil
}

func (ur *UserRepo) Save(ctx context.Context, user User) (int, error) {
	query := "INSERT INTO `users` (`username`, `password`, `role`, `created_at`, `updated_at`) VALUES (?,?,?,?,?)"

	var lastInsertId int64

	createdAt, updatedAt := time.Now(), time.Now()

	if !user.CreatedAt.IsZero() {
		createdAt = user.CreatedAt
	}

	if !user.UpdatedAt.IsZero() {
		updatedAt = user.UpdatedAt
	}

	result, err := ur.db.ExecContext(ctx, query,
		user.Username,
		user.Password,
		user.Role,
		createdAt.Format(time.DateTime),
		updatedAt.Format(time.DateTime))

	if err != nil {
		return int(lastInsertId), fmt.Errorf("failed to exec insert users query to db, error: %v", err)
	}

	lastInsertId, err = result.LastInsertId()

	if err != nil {
		return int(lastInsertId), fmt.Errorf("failed to get users last insert id after insert, error: %v", err)
	}

	_, err = result.RowsAffected()

	if err != nil {
		return int(lastInsertId), fmt.Errorf("user insert rows affected returns error: %v", err)
	}

	return int(lastInsertId), nil
}
