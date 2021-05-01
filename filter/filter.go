package filter

import (
	"fmt"
	"strings"
)

type Filter struct {
	// instreams  []string
	// content    string
	// outstreams []string
	opt *option
}

func New(opts ...OptionFunc) *Filter {
	o := &option{}
	for _, opt := range opts {
		opt(o)
	}
	return &Filter{
		opt: o,
	}
}

func (f *Filter) Params() (params []string) {
	if len(f.opt.instreams) != 0 {
		for _, stream := range f.opt.instreams {
			params = append(params, fmt.Sprintf("[%s]", stream))
		}
	}
	if f.opt.content != "" {
		params = append(params, f.opt.content)
	}
	if len(f.opt.outstreams) != 0 {
		for _, stream := range f.opt.outstreams {
			params = append(params, fmt.Sprintf("[%s]", stream))
		}
	}
	return
}

func (f *Filter) String() string {
	return strings.Join(f.Params(), "")
}
