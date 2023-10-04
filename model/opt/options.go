package opt

type Options struct {
	Language  string
	DateRange int
	Limit     int
	Start     string
	End       string
}

func ExtractOptions(opts ...any) Options {
	var options Options

	for _, option := range opts {
		if v, ok := option.(*LanguageOption); ok {
			options.Language = v.Get()
		}

		if v, ok := option.(*DateRangeOption); ok {
			options.DateRange = v.Get()
		}

		if v, ok := option.(*LimitOption); ok {
			options.Limit = v.Get()
		}

		if v, ok := option.(*StartOption); ok {
			options.Start = v.Get()
		}

		if v, ok := option.(*EndOption); ok {
			options.End = v.Get()
		}
	}

	return options
}
