package cmd

import (
	"testing"
	"time"
)

func TestParseEndOption(t *testing.T) {
	end := "         "
	end, err := parseEndDateTimeOption(end)

	if err != nil {
		t.Error(err)
	}

	if end != "" {
		t.Errorf("expected empty string but got: %s", end)
	}

	end = "2000-12-31 01:20:59"
	end, err = parseEndDateTimeOption(end)
	if err != nil {
		t.Error(err)
	}

	if end != "2000-12-31 01:20:59" {
		t.Errorf("unexpected time string: %s", end)
	}

	end = "2s"
	_, err = parseEndDateTimeOption(end)
	if err == nil {
		t.Error("expected invalid parse error but got nil")
	}

	end = "!2d"
	_, err = parseEndDateTimeOption(end)
	if err == nil {
		t.Error("expected invalid parse error but got nil")
	}

	end = "-2d"
	now := time.Now()
	end, err = parseEndDateTimeOption(end)
	if err != nil {
		t.Error(err)
	}

	endDate, err := time.Parse("2006-01-02 15:04:05", end)
	if err != nil {
		t.Error(err)
	}

	if now.Add(time.Duration(-2) * 24 * time.Hour).After(endDate) {
		t.Errorf("unepexted now: %v, end date: %v", now, endDate)
	}

	end = "2h"
	now = time.Now()
	end, err = parseEndDateTimeOption(end)
	if err != nil {
		t.Error(err)
	}

	endDate, err = time.Parse("2006-01-02 15:04:05", end)
	if err != nil {
		t.Error(err)
	}

	if now.Add(time.Duration(-2) * time.Hour).After(endDate) {
		t.Errorf("unepexted now: %v, end date: %v", now, endDate)
	}
}
