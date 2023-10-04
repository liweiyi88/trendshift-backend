package opt

import "strings"

type EndOption struct {
	value string
}

func End(value string) *EndOption {
	return &EndOption{value}
}

func (s *EndOption) Get() string {
	if s == nil {
		return ""
	}

	return strings.TrimSpace(s.value)
}
