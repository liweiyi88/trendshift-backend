package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liweiyi88/gti/search"
	"golang.org/x/exp/slog"
)

type SearchController struct {
	search search.Search
}

func NewSearchController() *SearchController {
	return &SearchController{
		search: search.NewSearch(),
	}
}

func (search *SearchController) Search(c *gin.Context) {
	results, err := search.search.SearchRepositories(c.Query("q"))

	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Error"})
		return
	}

	c.JSON(http.StatusOK, results)
}
