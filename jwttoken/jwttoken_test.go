package jwttoken

import (
	"testing"

	"github.com/liweiyi88/gti/model"
)

func TestGenerate(t *testing.T) {
	svc := NewTokenService("abcdefg")

	user := model.User{
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

	user := model.User{
		Id:       111,
		Username: "liweiyi88",
		Password: "testpass",
		Role:     "user,admin",
	}

	tokenString, err := svc.Generate(user)

	if err != nil {
		t.Error(err)
	}

	token, err := svc.Verify(tokenString)
	if err != nil {
		t.Error(err)
	}

	if claims, ok := token.Claims.(*AppClaim); ok && token.Valid {
		expcts := []struct {
			actual any
			want   any
		}{
			{
				actual: claims.UserId,
				want:   user.Id,
			},
			{
				actual: claims.Issuer,
				want:   "gti",
			},
			{
				actual: claims.Role,
				want:   user.Role,
			},
			{
				actual: claims.Subject,
				want:   user.Username,
			},
		}

		for _, test := range expcts {
			if test.actual != test.want {
				t.Errorf("expect: %v, actual got: %v", test.want, test.actual)
			}
		}
	} else {
		t.Error("invalid claims")
	}
}
