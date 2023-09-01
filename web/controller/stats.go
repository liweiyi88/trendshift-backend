package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liweiyi88/gti/model"
	"golang.org/x/exp/slog"
)

type StatsController struct {
	sr *model.StatsRepo
}

func NewStatsController(sr *model.StatsRepo) *StatsController {
	return &StatsController{
		sr: sr,
	}
}

func (sc *StatsController) GetDailyStats(c *gin.Context) {
	stats, err := sc.sr.FindDailyStats(c)

	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Error"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
