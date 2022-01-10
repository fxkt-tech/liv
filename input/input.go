package input

import (
	"strconv"
	"strings"
)

// Input is common input info.
type Input struct {
	opt *option
}

func New(opts ...OptionFunc) *Input {
	o := &option{}
	for _, opt := range opts {
		opt(o)
	}
	return &Input{
		opt: o,
	}
}

func (i *Input) Params() (params []string) {
	if i.opt.ss != 0 {
		params = append(params, "-ss", strconv.FormatFloat(i.opt.ss, 'f', -1, 64))
	}
	if i.opt.t != 0 {
		params = append(params, "-t", strconv.FormatFloat(i.opt.t, 'f', -1, 64))
	}
	params = append(params, "-i", i.opt.i)
	return
}

func (i *Input) String() string {
	return strings.Join(i.Params(), " ")
}
