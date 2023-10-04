package opt

type DateRangeOption struct {
	value int
}

func (d *DateRangeOption) Get() int {
	if d == nil {
		return 0
	}

	return d.value
}

func DateRange(value int) *DateRangeOption {
	return &DateRangeOption{value}
}
