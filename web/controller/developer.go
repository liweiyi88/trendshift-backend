package controller

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/liweiyi88/trendshift-backend/model"
	"github.com/liweiyi88/trendshift-backend/model/opt"
)

type DeveloperController struct {
	dr *model.DeveloperRepo
}

func NewDeveloperController(dr *model.DeveloperRepo) *DeveloperController {
	return &DeveloperController{dr}
}

func (dc *DeveloperController) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	repository, err := dc.dr.FindById(c, id) 

	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Error"})
		return
	}

	c.JSON(http.StatusOK, repository)
}

func (dc *DeveloperController) GetTrendingDevelopers(c *gin.Context) {
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

	developers, err := dc.dr.FindTrendingDevelopers(
		c,
		opt.Language(language),
		opt.Limit(limit),
		opt.DateRange(dateRange),
	)

	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Error"})
		return
	}

	c.JSON(http.StatusOK, developers)
}
