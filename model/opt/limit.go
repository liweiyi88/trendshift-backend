package opt

type LimitOption struct {
	value int
}

func (l *LimitOption) Get() int {
	if l == nil {
		return 0
	}

	return l.value
}

func Limit(value int) *LimitOption {
	return &LimitOption{value}
}
