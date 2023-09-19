package controller

import (
	"database/sql"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/liweiyi88/gti/github"
	"github.com/liweiyi88/gti/model"
	"golang.org/x/exp/slog"
)

type RepositoryController struct {
	grr      *model.GhRepositoryRepo
	ghClient *github.Client
}

type AttachTagsRequest struct {
	Id   int    `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

func NewRepositoryController(grr *model.GhRepositoryRepo, ghClient *github.Client) *RepositoryController {
	return &RepositoryController{
		grr:      grr,
		ghClient: ghClient,
	}
}

func (rc *RepositoryController) List(c *gin.Context) {
	ghRepositories, err := rc.grr.FindAllWithTags(c, c.Query("q"))
	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Error"})
		return
	}

	c.JSON(http.StatusOK, ghRepositories)
}

func (rc *RepositoryController) GetTrendingRepositories(c *gin.Context) {
	language, _ := url.QueryUnescape(c.Query("language"))
	limitQuery, _ := url.QueryUnescape(c.Query("limit"))
	dateRangeQuery := c.Query("range")

	var limit int
	var dateRange int
	var err error

	if dateRangeQuery != "" {
		dateRange, err = strconv.Atoi(dateRangeQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
			return
		}
	}

	if limitQuery != "" {
		limit, err = strconv.Atoi(limitQuery)

		if err != nil {
			slog.Error(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
			return
		}
	}

	repositories, err := rc.grr.FindTrendingRepositories(c, language, limit, dateRange)

	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Error"})
		return
	}

	c.JSON(http.StatusOK, repositories)
}

func (rc *RepositoryController) Get(c *gin.Context) {
	name, _ := url.QueryUnescape(c.Param("name"))

	repository, err := rc.grr.FindByName(c, name)

	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Error"})
		return
	}

	c.JSON(http.StatusOK, repository)
}

func (rc *RepositoryController) SaveTags(c *gin.Context) {
	repositoryId, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		slog.Error(err.Error())
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	tags := make([]model.Tag, 0, len(requestTags))

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
