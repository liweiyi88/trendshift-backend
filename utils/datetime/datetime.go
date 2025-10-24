package datetime

import (
	"context"
	"time"
)

type DateRange struct {
	StartDate time.Time
	EndDate   time.Time
}

func SleepWithContext(ctx context.Context, duration time.Duration) error {
	select {
	case <-time.After(duration):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func NewDateRange(start time.Time, end time.Time) *DateRange {
	return &DateRange{
		StartDate: start,
		EndDate:   end,
	}
}

func StartOfTomorrow() time.Time {
	return StartOfDay(time.Now().AddDate(0, 0, 1))
}

func StartOfToday() time.Time {
	return StartOfDay(time.Now())
}

func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func StartOfThisMonth() time.Time {
	return StartOfMonth(time.Now())
}

func EndOfThisMonth() time.Time {
	return EndOfMonth(time.Now())
}

func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func EndOfMonth(t time.Time) time.Time {
	firstOfNextMonth := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
	return firstOfNextMonth.Add(-time.Nanosecond)
}
