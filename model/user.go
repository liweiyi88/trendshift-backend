package model

import (
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       int
	Username string
	Password string
	Role     string
}

func (user *User) isPasswordValid(plainPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(plainPassword)) == nil
}

func (user *User) setPassword(plainPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)

	return nil
}
