package ffmpeg

import "fmt"

type Filter struct {
	alias   string
	content string
	others  []string
}

func NewFilter(alias string, content string, others ...string) *Filter {
	return &Filter{
		alias:   alias,
		content: content,
		others:  others,
	}
}

//
func (f *Filter) String() string {
	var steamsAlias string
	for _, other := range f.others {
		steamsAlias = fmt.Sprintf("%s[%s]", steamsAlias, other)
	}
	var alias string
	if f.alias != "" {
		alias = fmt.Sprintf("[%s]", f.alias)
	}
	return fmt.Sprintf("%s%s%s", steamsAlias, f.content, alias)
}
