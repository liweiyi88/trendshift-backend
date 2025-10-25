package controller

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/liweiyi88/trendshift-backend/model"
	"github.com/liweiyi88/trendshift-backend/utils/datetime"
)

type RepositoryEngagementController struct {
	rmr *model.RepositoryMonthlyInsightRepo
}

func NewRepositoryEngagementController(rmr *model.RepositoryMonthlyInsightRepo) *RepositoryEngagementController {
	return &RepositoryEngagementController{
		rmr: rmr,
	}
}

// Valid query parameter example: ?year=2025&month=10&language=Go&limit=10&created_after=2024-01-02T15:04:05+10:00
func (controller *RepositoryEngagementController) List(c *gin.Context) {
	ts := datetime.StartOfThisMonth()
	defaultYear, defaultMonth := strconv.Itoa(ts.Year()), strconv.Itoa(int(ts.Month()))

	yearStr := c.DefaultQuery("year", defaultYear)
	monthStr := c.DefaultQuery("month", defaultMonth)
	languageStr := c.DefaultQuery("language", "")
	createdAfterStr := c.DefaultQuery("created_after", "")
	limitStr := c.DefaultQuery("limit", "10")

	params, err := model.NewListEngagementParams(c.Param("metric"), yearStr, monthStr, languageStr, limitStr, createdAfterStr)
	if err != nil {
		slog.Error("params are not valid", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	data, err := controller.rmr.FindRepositoryMonthlyEngagements(c, params)
	if err != nil {
		slog.Error("failed to fetch repository engagements", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, data)
}
