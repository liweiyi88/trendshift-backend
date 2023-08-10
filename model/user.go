package model

import (
	"time"

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
