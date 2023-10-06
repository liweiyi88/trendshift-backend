package controller

import (
	"net/http"
	"strconv"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/liweiyi88/gti/model"
)

type StatsController struct {
	sr *model.StatsRepo
}

func NewStatsController(sr *model.StatsRepo) *StatsController {
	return &StatsController{
		sr: sr,
	}
}

func (sc *StatsController) GetTrendingTopicsStats(c *gin.Context) {
	dateRange := c.Query("range")

	var r int
	var err error

	if dateRange != "" {
		r, err = strconv.Atoi(dateRange)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
			return
		}
	}

	stats, err := sc.sr.FindTrendingTopicsStats(c, r)

	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Error"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
