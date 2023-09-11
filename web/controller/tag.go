package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/liweiyi88/gti/model"
	"golang.org/x/exp/slog"
)

type TagController struct {
	tr *model.TagRepo
}

type CreateTagRequest struct {
	Name string `json:"name" binding:"required"`
}

func NewTagController(tr *model.TagRepo) *TagController {
	return &TagController{
		tr: tr,
	}
}

func (tc *TagController) List(c *gin.Context) {
	name := c.Query("name")

	tags, err := tc.tr.Find(c, name)

	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Error"})
		return
	}

	c.JSON(http.StatusOK, tags)
}

func (tc *TagController) Save(c *gin.Context) {
	var request CreateTagRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := tc.tr.FindByName(c, strings.ToLower(request.Name))

	if err == nil {
		c.JSON(http.StatusCreated, tag)
		return
	}

	tag.Name = request.Name
	id, err := tc.tr.Save(c, tag)
	tag.Id = int(id)

	if err != nil {
		slog.Error(err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Error"})
		return
	}

	c.JSON(http.StatusCreated, tag)
}
