package jwttoken

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/liweiyi88/gti/trending"
)

type TokenService struct {
	signingKey []byte
}

func NewTokenService(signingKey string) *TokenService {
	return &TokenService{
		signingKey: []byte(signingKey),
	}
}

func (t *TokenService) Generate(user trending.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.Username,
		"iss":  "gti",
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
		"role": strings.Join(user.Role, ","),
	})

	return token.SignedString(t.signingKey)
}

func (t *TokenService) Verify(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return t.signingKey, nil
	})

	return token, err
}
