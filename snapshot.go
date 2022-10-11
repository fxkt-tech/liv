package liv

import (
	"context"

	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/filter"
	"github.com/fxkt-tech/liv/ffmpeg/input"
	"github.com/fxkt-tech/liv/ffmpeg/naming"
	"github.com/fxkt-tech/liv/ffmpeg/output"
	"github.com/fxkt-tech/liv/internal/math"
)

type Snapshot struct {
	*options

	spec *SnapshotSpec
}

func NewSnapshot(opts ...Option) *Snapshot {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	ss := &Snapshot{
		spec:    NewSnapshotSpec(),
		options: o,
	}
	return ss
}

func (ss *Snapshot) Simple(ctx context.Context, params *SnapshotParams) error {
	err := ss.spec.CheckSatified(params)
	if err != nil {
		return err
	}

	var (
		nm            = naming.New()
		inputs        input.Inputs
		filters       filter.Filters
		outputOptions []output.OutputOption
	)

	inputs = append(inputs, input.WithTime(params.StartTime, 0, params.Infile))

	// 使用普通帧截图时，必须要传截图间隔，除非只截一张
	switch params.FrameType {
	case 0: // 关键帧
		filters = append(filters, filter.Select(nm.Empty(), "'eq(pict_type,I)'"))
		outputOptions = append(outputOptions, output.VSync("vfr"))
	case 1:
		if params.Num != 1 {
			filters = append(filters, filter.FPS(nm.Empty(), math.Fraction(1, params.Interval)))
		}
	}

	outputOptions = append(outputOptions,
		output.Vframes(params.Num),
		output.File(params.Outfile),
	)

	return ffmpeg.NewFFmpeg(ss.ffmpegOpts...).
		AddInput(inputs...).
		AddFilter(filters...).
		AddOutput(output.New(outputOptions...)).
		Run(ctx)
}
