package controller

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

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

// Valid query parameter example: ?year=2024&month=1&language=PHP&createdAt=2026-01-02T15:04:05+10:00
func (controller *RepositoryEngagementController) List(c *gin.Context) {
	defaultYear, defaultMonth := strconv.Itoa(datetime.StartOfThisMonth().Year()), datetime.StartOfThisMonth().Month().String()

	year := c.DefaultQuery("year", defaultYear)
	month := c.DefaultQuery("month", defaultMonth)
	language := c.DefaultQuery("language", "")
	createdAfterStr := c.DefaultQuery("createdAfter", "")
	limitStr := c.DefaultQuery("limit", "10")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request, invalid limit"})
		return
	}

	var createdAfter time.Time
	if createdAfterStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, createdAfterStr)
		if err == nil {
			createdAfter = parsedTime
		}
	}

	params := model.ListEngagementParams{
		Metric:       c.Param("metric"),
		Year:         year,
		Month:        month,
		Language:     language,
		Limit:        limit,
		CreatedAfter: createdAfter,
	}

	if err := params.Validate(); err != nil {
		slog.Error("params are not valid", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	data, err := controller.rmr.FindRepositoryMonthlyEngagements(c, params)
	if err != nil {
		slog.Error("failed to fetch engagements", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, data)
}
