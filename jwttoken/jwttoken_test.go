package jwttoken

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/liweiyi88/gti/trending"
)

func TestGenerate(t *testing.T) {
	svc := NewTokenService("abcdefg")

	user := trending.User{
		Username: "liweiyi88",
		Password: "testpass",
	}

	token, err := svc.Generate(user)

	if err != nil {
		t.Error(err)
	}

	t.Log(token)
}

func TestVerify(t *testing.T) {
	svc := NewTokenService("abcdefg")

	user := trending.User{
		Username: "liweiyi88",
		Password: "testpass",
		Role:     []string{"user", "admin"},
	}

	tokenString, err := svc.Generate(user)

	if err != nil {
		t.Error(err)
	}

	token, err := svc.Verify(tokenString)
	if err != nil {
		t.Error(err)
	}

	expect := make(map[string]any, 0)
	expect["iss"] = "gti"
	expect["role"] = "user,admin"
	expect["sub"] = "liweiyi88"

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		for key, value := range expect {
			if claims[key] != value {
				t.Errorf("expect claim %s has value %v, but got %v", key, claims[key], value)
			}
		}
	} else {
		t.Error("invalid claims")
	}
}
