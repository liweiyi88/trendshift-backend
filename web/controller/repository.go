package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/liweiyi88/gti/model"
	"golang.org/x/exp/slog"
)

type RepositoryController struct {
	grr *model.GhRepositoryRepo
}

type AttachTagsRequest struct {
	Id   int    `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

func NewRepositoryController(grr *model.GhRepositoryRepo) *RepositoryController {
	return &RepositoryController{
		grr: grr,
	}
}

func (rc *RepositoryController) List(c *gin.Context) {
	ghRepositories, err := rc.grr.FindAllWithTags(c)
	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Error"})
		return
	}

	c.JSON(http.StatusOK, ghRepositories)
}

func (rc *RepositoryController) SaveTags(c *gin.Context) {
	repositoryId, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusNotFound, nil)
		return
	}

	ghRepository, err := rc.grr.FindById(c, repositoryId)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, nil)
			return
		}

		slog.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Error"})
	}

	requestTags := make([]AttachTagsRequest, 0)

	if err := c.ShouldBindJSON(&requestTags); err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tags := make([]model.Tag, len(requestTags))

	for _, rt := range requestTags {
		var tag model.Tag

		tag.Id = rt.Id
		tag.Name = rt.Name

		tags = append(tags, tag)
	}

	err = rc.grr.SaveTags(c, ghRepository, tags)

	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Error"})
		return
	}

	c.JSON(http.StatusOK, ghRepository)
}
