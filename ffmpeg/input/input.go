package input

import (
	"fmt"
	"strings"
)

type InputOption func(*Input)

func I(i string) InputOption {
	return func(o *Input) {
		o.i = i
	}
}

func StartTime(ss float64) InputOption {
	return func(o *Input) {
		o.ss = ss
	}
}

func Duration(t float64) InputOption {
	return func(o *Input) {
		o.t = t
	}
}

// Input is common input info.
type Input struct {
	i  string  // i is input file.
	ss float64 // ss is start_time.
	t  float64 // t is duration.
	// ext []string // extra params.
}

func New(opts ...InputOption) *Input {
	i := &Input{}
	for _, o := range opts {
		o(i)
	}
	return i
}

func WithSimple(i string) *Input {
	return &Input{i: i}
}

func WithTime(ss, t float64, i string) *Input {
	return &Input{
		ss: ss,
		t:  t,
		i:  i,
	}
}

func (i *Input) Params() (params []string) {
	if i.ss != 0 {
		params = append(params, "-ss", fmt.Sprintf("%.2f", i.ss))
	}
	if i.t != 0 {
		params = append(params, "-t", fmt.Sprintf("%.2f", i.t))
	}
	params = append(params, "-i", i.i)
	return
}

func (i *Input) String() string {
	return strings.Join(i.Params(), " ")
}

//

type Inputs []*Input

func (inputs Inputs) Params() (params []string) {
	for _, input := range inputs {
		params = append(params, input.Params()...)
	}
	return
}

func (inputs Inputs) String() string {
	return strings.Join(inputs.Params(), " ")
}