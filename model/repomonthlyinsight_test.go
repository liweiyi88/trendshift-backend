package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListEngagementParamsValidate(t *testing.T) {
	params := ListEngagementParams{
		Metric:   "invalid",
		Language: "PHP",
	}
	assert.Error(t, params.Validate())

	metrics := []string{"stars", "forks", "merged_prs", "issues", "closed_issues"}
	for _, m := range metrics {
		t.Run(m, func(t *testing.T) {
			params.Metric = m
			assert.NoError(t, params.Validate())
		})
	}
}
