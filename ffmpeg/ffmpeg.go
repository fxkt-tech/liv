package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/fxkt-tech/liv/ffmpeg/filter"
	"github.com/fxkt-tech/liv/ffmpeg/input"
	"github.com/fxkt-tech/liv/ffmpeg/output"
	"github.com/fxkt-tech/liv/internal/encoding/json"
	"github.com/fxkt-tech/liv/internal/sugar"
)

type FFmpeg struct {
	dry bool // dry run

	debug bool

	bin string
	v   LogLevel // v is log level.
	y   bool     // y is yes for overwrite output file.

	inputs  input.Inputs
	filters filter.Filters
	outputs output.Outputs
}

func New(opts ...Option) *FFmpeg {
	fm := &FFmpeg{
		bin: "ffmpeg",
		y:   true,
		v:   LogLevelError,
	}
	sugar.Range(opts, func(o Option) { o(fm) })
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

	params = append(params, fm.inputs.Tidy().Params()...)
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
	if ff.debug {
		ff.DryRun()
	} else {
		if ff.dry {
			ff.DryRun()
			return nil
		}
	}
	cc := exec.CommandContext(ctx, ff.bin, ff.Params()...)
	retbytes, err := cc.CombinedOutput()
	if err != nil {
		err = errors.New(string(retbytes))
	}
	return
}

type LoudnormParms struct {
	InputI            float32 `json:"input_i,string"`
	InputTP           float32 `json:"input_tp,string"`
	InputLRA          float32 `json:"input_lra,string"`
	InputThresh       float32 `json:"input_thresh,string"`
	OutputI           float32 `json:"output_i,string"`
	OutputTP          float32 `json:"output_tp,string"`
	OutputLRA         float32 `json:"output_lra,string"`
	OutputThresh      float32 `json:"output_thresh,string"`
	NormalizationType string  `json:"normalization_type,string"`
	TargetOffset      float32 `json:"target_offset,string"`
}

func (ff *FFmpeg) ExtractLoudnorm(ctx context.Context) (*LoudnormParms, error) {
	if ff.debug {
		ff.DryRun()
	}
	cc := exec.CommandContext(ctx, ff.bin, ff.Params()...)
	retbytes, err := cc.CombinedOutput()
	retstr := string(retbytes)
	if err != nil {
		return nil, errors.New(retstr)
	}

	reg := regexp.MustCompile("(?s)({.*})")
	matches := reg.FindStringSubmatch(retstr)
	if len(matches) < 2 {
		return nil, errors.New(retstr)
	}
	return json.ToV[*LoudnormParms]([]byte(matches[1])), nil
}
