package input

import (
	"strconv"
	"strings"
)

// Input is common input info.
type Input struct {
	// i   string   // i is input file.
	// ss  int64    // ss is starttime.
	// t   int64    // t is duration.
	// ext []string // extra params.
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
		params = append(params, "-ss", strconv.FormatInt(i.opt.ss, 10))
	}
	if i.opt.t != 0 {
		params = append(params, "-t", strconv.FormatInt(i.opt.t, 10))
	}
	params = append(params, "-i", i.opt.i)
	return
}

func (i *Input) String() string {
	return strings.Join(i.Params(), " ")
}
