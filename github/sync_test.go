package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchTotalContributors(t *testing.T) {
	total := fetchTotalContributors("liweiyi88/onedump")
	assert.Equal(t, 3, total)
}
