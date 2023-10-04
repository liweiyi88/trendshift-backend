package opt

import "strings"

type LanguageOption struct {
	value string
}

func Language(value string) *LanguageOption {
	return &LanguageOption{value}
}

func (l *LanguageOption) Get() string {
	if l == nil {
		return ""
	}

	return strings.TrimSpace(l.value)
}
