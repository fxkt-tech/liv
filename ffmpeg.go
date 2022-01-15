package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"fxkt.tech/ffmpeg/filter"
	"fxkt.tech/ffmpeg/input"
	"fxkt.tech/ffmpeg/output"
)

type FFmpeg struct {
	cmd     string
	v       string // v is log level.
	y       bool   // y is yes for overwrite output file.
	inputs  input.Inputs
	filters filter.Filters
	outputs output.Outputs
	// Sentence string
	opt *option
}

func Default() *FFmpeg {
	return &FFmpeg{
		cmd: "ffmpeg",
		y:   true,
		opt: &option{},
	}
}

func New(opts ...OptionFunc) *FFmpeg {
	o := &option{}
	for _, opt := range opts {
		opt(o)
	}
	return &FFmpeg{
		cmd: "ffmpeg",
		y:   true,
		opt: o,
	}
}

func (ff *FFmpeg) CmdLoc(loc string) {
	ff.cmd = loc
}

func (ff *FFmpeg) Yes(y bool) {
	ff.y = y
}

func (ff *FFmpeg) LogLevel(v string) {
	ff.v = v
}

func (ff *FFmpeg) AddInput(inputs ...*input.Input) {
	ff.inputs = append(ff.inputs, inputs...)
}

func (ff *FFmpeg) AddFilter(filters ...*filter.Filter) {
	ff.filters = append(ff.filters, filters...)
}

func (ff *FFmpeg) AddOutput(outputs ...*output.Output) {
	ff.outputs = append(ff.outputs, outputs...)
}

func (ff *FFmpeg) Params() (params []string) {
	if ff.v != "" {
		params = append(params, "-v", ff.v)
	}
	if ff.y {
		params = append(params, "-y")
	}
	params = append(params, ff.inputs.Params()...)
	params = append(params, ff.filters.Params()...)
	params = append(params, ff.outputs.Params()...)
	return
}

func (ff *FFmpeg) DryRun() {
	var ps []string
	ps = append(ps, ff.cmd)
	ps = append(ps, ff.Params()...)
	fmt.Println(strings.Join(ps, " "))
}

func (ff *FFmpeg) Run(ctx context.Context) (err error) {
	cc := exec.CommandContext(ctx, ff.cmd, ff.Params()...)
	retbytes, err := cc.CombinedOutput()
	if err != nil {
		errstr := strings.ReplaceAll(string(retbytes), "\n", "|")
		if len(errstr) > 200 {
			errstr = errstr[:200]
		}
		err = errors.New(errstr)
	}
	return
}
