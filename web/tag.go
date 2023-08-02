package web

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liweiyi88/gti/trending"
)

func postTag(c *gin.Context) {
	var tag trending.Tag

	if err := c.ShouldBind(&tag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Println(tag)

	c.JSON(http.StatusCreated, tag)
}
