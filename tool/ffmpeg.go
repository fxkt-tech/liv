package echotool

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"fxkt.tech/echo/tool/filter"
	"fxkt.tech/echo/tool/input"
	"fxkt.tech/echo/tool/output"
)

type FFmpegOption func(*FFmpeg)

func CmdLoc(loc string) FFmpegOption {
	return func(fm *FFmpeg) {
		fm.cmd = loc
	}
}

func Yes(y bool) FFmpegOption {
	return func(fm *FFmpeg) {
		fm.y = y
	}
}

func LogLevel(v string) FFmpegOption {
	return func(fm *FFmpeg) {
		fm.v = v
	}
}

type FFmpeg struct {
	cmd string
	v   string // v is log level.
	y   bool   // y is yes for overwrite output file.

	inputs  input.Inputs
	filters filter.Filters
	outputs output.Outputs
}

type Generater interface {
	String() string
}

func NewFFmpeg(opts ...FFmpegOption) *FFmpeg {
	fm := &FFmpeg{
		cmd: "ffmpeg",
		y:   true,
		v:   "error",
	}
	for _, o := range opts {
		o(fm)
	}
	return fm
}

func (fm *FFmpeg) AddInput(inputs ...*input.Input) *FFmpeg {
	fm.inputs = append(fm.inputs, inputs...)
	return fm
}

func (fm *FFmpeg) AddFilter(filters ...filter.Filter) *FFmpeg {
	fm.filters = append(fm.filters, filters...)
	return fm
}

func (fm *FFmpeg) AddOutput(outputs ...*output.Output) *FFmpeg {
	fm.outputs = append(fm.outputs, outputs...)
	return fm
}

func (fm *FFmpeg) Params() []string {
	var params []string
	if fm.v != "" {
		params = append(params, "-v", fm.v)
	}
	if fm.y {
		params = append(params, "-y")
	}
	params = append(params, fm.inputs.Params()...)
	params = append(params, fm.filters.Params()...)
	params = append(params, fm.outputs.Params()...)
	return params
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
