package opt

import "strings"

type StartOption struct {
	value string
}

func Start(value string) *StartOption {
	return &StartOption{value}
}

func (s *StartOption) Get() string {
	if s == nil {
		return ""
	}

	return strings.TrimSpace(s.value)
}
