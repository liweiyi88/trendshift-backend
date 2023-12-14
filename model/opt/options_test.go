package opt

import (
	"testing"
)

func TestExtractOptionsWithEmptyArgs(t *testing.T) {
	options := ExtractOptions()

	expcts := []struct {
		actual any
		want   any
	}{
		{
			actual: options.Language,
			want:   "",
		},
		{
			actual: options.DateRange,
			want:   0,
		},
		{
			actual: options.End,
			want:   "",
		},
		{
			actual: options.Start,
			want:   "",
		},
		{
			actual: options.Limit,
			want:   0,
		},
	}

	for _, test := range expcts {
		if test.actual != test.want {
			t.Errorf("expect: %v, actual got: %v", test.want, test.actual)
		}
	}
}

func TestExtractOptionsWithArgs(t *testing.T) {
	options := ExtractOptions(
		DateRange(7),
		Language("Go"),
		Start("2023-10-04 00:00:00"),
		End("2023-10-04 23:59:59"),
		Limit(24),
	)

	expcts := []struct {
		actual any
		want   any
	}{
		{
			actual: options.Language,
			want:   "Go",
		},
		{
			actual: options.DateRange,
			want:   7,
		},
		{
			actual: options.End,
			want:   "2023-10-04 23:59:59",
		},
		{
			actual: options.Start,
			want:   "2023-10-04 00:00:00",
		},
		{
			actual: options.Limit,
			want:   24,
		},
	}

	for _, test := range expcts {
		if test.actual != test.want {
			t.Errorf("expect: %v, actual got: %v", test.want, test.actual)
		}
	}
}

func TestMaxLimit(t *testing.T) {
	options := ExtractOptions(
		DateRange(7),
		Language("Go"),
		Start("2023-10-04 00:00:00"),
		End("2023-10-04 23:59:59"),
		Limit(101),
	)

	expcts := []struct {
		actual any
		want   any
	}{
		{
			actual: options.Limit,
			want:   100,
		},
	}

	for _, test := range expcts {
		if test.actual != test.want {
			t.Errorf("expect: %v, actual got: %v", test.want, test.actual)
		}
	}
}
