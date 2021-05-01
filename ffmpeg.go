package ffmpeg

import (
	"context"
	"log"
	"os/exec"

	"fxkt.tech/ffmpeg/filter"
	"fxkt.tech/ffmpeg/input"
	"fxkt.tech/ffmpeg/output"
)

type FFmpeg struct {
	cmd      string
	y        bool // y is yes for cover output file.
	inputs   input.Inputs
	filters  filter.Filters
	outputs  output.Outputs
	Sentence string
}

func Default() *FFmpeg {
	return &FFmpeg{
		cmd: "ffmpeg",
		y:   true,
	}
}

func (ff *FFmpeg) Yes(y bool) {
	ff.y = y
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
	if ff.y {
		params = append(params, "-y")
	}
	params = append(params, ff.inputs.Params()...)
	params = append(params, ff.filters.String())
	params = append(params, ff.outputs.Params()...)
	return
}

func (ff *FFmpeg) Run(ctx context.Context) (err error) {
	cc := exec.CommandContext(ctx, ff.cmd, ff.Params()...)
	ff.Sentence = cc.String()
	retbytes, err := cc.CombinedOutput()
	if err != nil {
		return
	}
	log.Println(string(retbytes))
	return
}
