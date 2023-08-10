package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liweiyi88/gti/config"
	"github.com/liweiyi88/gti/jwttoken"
	"github.com/liweiyi88/gti/model"
	"golang.org/x/exp/slog"
)

type SecurityController struct {
	ur *model.UserRepo
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type tokenResponse struct {
	Token string `json:"access_token"`
}

func NewSecurityController(ur *model.UserRepo) *SecurityController {
	return &SecurityController{
		ur: ur,
	}
}

func (sc *SecurityController) Login(c *gin.Context) {
	var request LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := sc.ur.FindByNamePassword(c, request.Username, request.Password)

	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}

	jwtsvc := jwttoken.NewTokenService(config.SignIngKey)
	tokenString, err := jwtsvc.Generate(user)

	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Error"})
		return
	}

	var response tokenResponse
	response.Token = tokenString
	c.JSON(http.StatusOK, response)
}
