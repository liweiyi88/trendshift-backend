package jwttoken

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/liweiyi88/gti/model"
)

type AppClaim struct {
	UserId int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenService struct {
	signingKey []byte
}

func NewTokenService(signingKey string) *TokenService {
	return &TokenService{
		signingKey: []byte(signingKey),
	}
}

func (t *TokenService) Generate(user model.User) (string, int64, error) {
	expiredAt := time.Now().Add(15 * time.Minute).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.Id,
		"sub":     user.Username,
		"iss":     "gti",
		"exp":     expiredAt,
		"role":    user.Role,
	})

	tokenString, err := token.SignedString(t.signingKey)

	return tokenString, expiredAt, err
}

func (t *TokenService) Verify(tokenString string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AppClaim{}, func(token *jwt.Token) (any, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return t.signingKey, nil
	})

	return token, err
}
