package opt

const maxLimit = 2000

type LimitOption struct {
	value int
}

func (l *LimitOption) Get() int {
	if l == nil {
		return 0
	}

	if l.value > maxLimit {
		return maxLimit
	}

	return l.value
}

func Limit(value int) *LimitOption {
	return &LimitOption{value}
}
