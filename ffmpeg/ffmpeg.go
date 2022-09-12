package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"fxkt.tech/liv/ffmpeg/filter"
	"fxkt.tech/liv/ffmpeg/input"
	"fxkt.tech/liv/ffmpeg/output"
)

type FFmpegOption func(*FFmpeg)

func Binary(bin string) FFmpegOption {
	return func(fm *FFmpeg) {
		fm.bin = bin
	}
}

func Yes(y bool) FFmpegOption {
	return func(fm *FFmpeg) {
		fm.y = y
	}
}

func V(v LogLevel) FFmpegOption {
	return func(fm *FFmpeg) {
		fm.v = v
	}
}

func Dry(dry bool) FFmpegOption {
	return func(f *FFmpeg) {
		f.dry = dry
	}
}

type FFmpeg struct {
	dry bool // dry run

	bin string
	v   LogLevel // v is log level.
	y   bool     // y is yes for overwrite output file.

	inputs  input.Inputs
	filters filter.Filters
	outputs output.Outputs
}

type Generater interface {
	String() string
}

func NewFFmpeg(opts ...FFmpegOption) *FFmpeg {
	fm := &FFmpeg{
		bin: "ffmpeg",
		y:   true,
		v:   LogLevelError,
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
		params = append(params, "-v", fm.v.String())
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
	ps = append(ps, ff.bin)
	ps = append(ps, ff.Params()...)
	fmt.Println(strings.Join(ps, " "))
}

func (ff *FFmpeg) Run(ctx context.Context) (err error) {
	if ff.dry {
		ff.DryRun()
		return nil
	}
	cc := exec.CommandContext(ctx, ff.bin, ff.Params()...)
	retbytes, err := cc.CombinedOutput()
	if err != nil {
		err = errors.New(string(retbytes))
	}
	return
}
