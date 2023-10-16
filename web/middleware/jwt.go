package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/liweiyi88/trendshift-backend/config"
	"github.com/liweiyi88/trendshift-backend/jwttoken"
)

func JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := verifyToken(c)

		if err != nil {
			slog.Error("authentication failed", slog.Any("error", err))
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		c.Next()
	}
}

func verifyToken(c *gin.Context) error {
	authHeader := c.Request.Header.Get("Authorization")

	bearerString := strings.Split(authHeader, " ")

	if len(bearerString) != 2 {
		return errors.New("incorrectly formatted authorization header")
	}

	tokenString := bearerString[1]

	token, err := jwttoken.NewTokenService(config.SignIngKey).Verify(tokenString)

	if err != nil {
		return err
	}

	_, ok := token.Claims.(*jwttoken.AppClaim)

	if ok && token.Valid {
		return nil
	} else {
		return fmt.Errorf("invalid token string: %v", tokenString)
	}
}
