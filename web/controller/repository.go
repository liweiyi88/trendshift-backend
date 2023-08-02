package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/liweiyi88/gti/trending"
	"golang.org/x/exp/slog"
)

type RepositoryController struct {
	grr *trending.GhRepositoryRepo
}

func NewRepositoryController(grr *trending.GhRepositoryRepo) *RepositoryController {
	return &RepositoryController{
		grr: grr,
	}
}

func (rc *RepositoryController) AttachTagsToRepository(c *gin.Context) {
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

	requestTags := make([]trending.UpdateTagRequest, 0)

	if err := c.ShouldBind(&requestTags); err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tags := make([]trending.Tag, 0)

	for _, rt := range requestTags {
		var tag trending.Tag

		tag.Id = rt.Id
		tag.Name = rt.Name

		tags = append(tags, tag)
	}

	err = rc.grr.UpdateWithTags(c, ghRepository, tags)

	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Error"})
		return
	}

	c.JSON(http.StatusOK, repositoryId)
}
