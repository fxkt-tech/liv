package liv

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	svg "github.com/ajstarks/svgo"
	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/filter"
	"github.com/fxkt-tech/liv/ffmpeg/input"
	"github.com/fxkt-tech/liv/ffmpeg/output"
	"github.com/fxkt-tech/liv/ffmpeg/stream"
	"github.com/fxkt-tech/liv/ffprobe"
	"github.com/fxkt-tech/liv/pkg/math"
	"github.com/fxkt-tech/liv/pkg/sugar"
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
		inputs        input.Inputs
		filters       filter.Filters
		outputOptions []output.Option
	)

	inputs = append(inputs, input.WithTime(params.StartTime, 0, params.Infile))

	lastFilter := stream.V(0)
	// 使用普通帧截图时，必须要传截图间隔，除非只截一张
	switch params.FrameType {
	case 0: // 仅关键帧截图
		selectFilter := filter.Select("'eq(pict_type,I)'")
		filters = append(filters, selectFilter)
		lastFilter = selectFilter
		outputOptions = append(outputOptions, output.VSync("vfr"))
	case 1: // 等间隔截图
		if params.Num != 1 {
			if params.IntervalFrames > 0 {
				selectFilter := filter.Select(fmt.Sprintf("'not(mod(n,%d))'", params.IntervalFrames))
				filters = append(filters, selectFilter)
				lastFilter = selectFilter
				outputOptions = append(outputOptions, output.VSync("vfr"))
			} else {
				fpsFilter := filter.FPS(math.Fraction(1, params.Interval))
				filters = append(filters, fpsFilter)
				lastFilter = fpsFilter
			}
		}
	case 2: // 指定帧序列截图
		if len(params.Frames) > 0 {
			selectExpr := fmt.Sprintf("'%s'",
				strings.Join(sugar.Map(params.Frames, func(frame int32) string {
					return fmt.Sprintf("eq(n,%d)", frame)
				}), "+"))
			selectFilter := filter.Select(selectExpr)
			filters = append(filters, selectFilter)
			lastFilter = selectFilter
			outputOptions = append(outputOptions, output.VSync("vfr"))
		}
	}
	if params.Width > 0 || params.Height > 0 {
		scaleFilter := filter.Scale(params.Width, params.Height).Use(lastFilter)
		filters = append(filters, scaleFilter)
		lastFilter = scaleFilter
	}

	if len(filters) > 0 {
		outputOptions = append(outputOptions, output.Map(lastFilter))
	}
	outputOptions = append(outputOptions,
		output.Vframes(params.Num),
		output.Format("image2"),
		output.File(params.Outfile),
	)

	return ffmpeg.New(ss.ffmpegOpts...).
		AddInput(inputs...).
		AddFilter(filters...).
		AddOutput(output.New(outputOptions...)).
		Run(ctx)
}

func (ss *Snapshot) Sprite(ctx context.Context, params *SpriteParams) error {
	var (
		duration = float32(params.XLen*params.YLen) * params.Interval
		frames   = params.XLen * params.YLen
	)

	// 获取视频信息
	ffp := ffprobe.New(ss.ffprobeOpts...).
		Input(params.Infile)
	err := ffp.Run(ctx)
	if err != nil {
		return err
	}

	// 获取视频实际时长和帧数
	vstream := ffp.GetFirstVideoStream()
	if vstream == nil {
		return errors.New("vstream is nil")
	}
	// 若视频实际帧数少于预期，则根据实际帧数生成雪碧图
	if vstream.NBFrames > 0 && vstream.NBFrames < frames {
		frames = vstream.NBFrames
	}
	if vstream.Duration > 0 {
		duration = vstream.Duration
	}

	var (
		inputs        input.Inputs
		filters       filter.Filters
		outputOptions []output.Option
	)

	inputs = append(inputs, input.WithSimple(params.Infile))

	filterFPS := filter.FPS(math.Fraction(frames, duration))
	filterScale := filter.Scale(params.Width, params.Height).Use(filterFPS)
	filterTile := filter.Tile(params.XLen, params.YLen).Use(filterScale)

	filters = append(filters, filterFPS, filterScale, filterTile)
	outputOptions = append(outputOptions,
		output.Map(filterTile.S()),
		output.File(params.Outfile),
	)

	return ffmpeg.New(ss.ffmpegOpts...).
		AddInput(inputs...).
		AddFilter(filters...).
		AddOutput(output.New(outputOptions...)).
		Run(ctx)
}

func (ss *Snapshot) SVGMark(ctx context.Context, params *SVGMarkParams) error {
	// 初始化
	var (
		outfolder         = filepath.Dir(params.Infile)
		snapshotlocalfile = fmt.Sprintf("%s/ss.jpg", outfolder)
		svglocalfile      = fmt.Sprintf("%s/pz.svg", outfolder)
	)

	// 截图
	err := ss.Simple(ctx, &SnapshotParams{
		Infile:    params.Infile,
		Outfile:   snapshotlocalfile,
		StartTime: params.StartTime,
		Num:       1,
		FrameType: 1,
	})
	if err != nil {
		return err
	}

	ffp := ffprobe.New(ss.ffprobeOpts...).Input(snapshotlocalfile)
	err = ffp.Run(ctx)
	if err != nil {
		return err
	}
	vstream := ffp.GetFirstVideoStream()
	if vstream == nil {
		return errors.New("file has not video stream")
	}

	f, err := os.Create(svglocalfile)
	if err != nil {
		return err
	}
	canvas := svg.New(f)
	canvas.Start(int(vstream.Width), int(vstream.Height))
	canvas.Image(0, 0, int(vstream.Width), int(vstream.Height), snapshotlocalfile)
	for _, annotation := range params.Annotations {
		switch annotation.Type {
		case "rect":
			fromx := int(annotation.FromPoint.X * float32(vstream.Width))
			fromy := int(annotation.FromPoint.Y * float32(vstream.Height))
			tox := int(annotation.ToPoint.X * float32(vstream.Width))
			toy := int(annotation.ToPoint.Y * float32(vstream.Height))
			minx := int(min(float32(fromx), float32(tox)))
			miny := int(min(float32(fromy), float32(toy)))
			w := int(math.Abs(float32(fromx - tox)))
			h := int(math.Abs(float32(fromy - toy)))
			styles := []string{"fill:transparent"}
			if annotation.Stroke != "" {
				styles = append(styles, fmt.Sprintf("stroke:%s", annotation.Stroke))
			}
			if annotation.StrokeWidth != 0 {
				styles = append(styles, fmt.Sprintf("stroke-width:%dpx", annotation.StrokeWidth))
			}
			canvas.Rect(minx, miny, w, h, strings.Join(styles, ";"))
		case "pen":
			var d string
			plen := len(annotation.Points)
			for i, point := range annotation.Points {
				x := int(point.X * float32(vstream.Width))
				y := int(point.Y * float32(vstream.Height))
				if i == 0 {
					d = fmt.Sprintf("%sM%d %d ", d, x, y)
				} else if i == plen-1 {
					d = fmt.Sprintf("%sL%d %d", d, x, y)
				} else {
					d = fmt.Sprintf("%sL%d %d ", d, x, y)
				}
			}
			styles := []string{"fill:transparent"}
			if annotation.Stroke != "" {
				styles = append(styles, fmt.Sprintf("stroke:%s", annotation.Stroke))
			}
			if annotation.StrokeWidth != 0 {
				styles = append(styles, fmt.Sprintf("stroke-width:%dpx", annotation.StrokeWidth))
			}
			canvas.Path(d, strings.Join(styles, ";"))
		case "arrow":
			// 单位箭头： M 0 0 L 0.8 0.04 L 0.8 0.08 L 1 0 L 0.8 -0.08 L 0.8 -0.04 Z
			unitPoints := []*Point{
				{X: 0, Y: 0},
				{X: 0.8, Y: 0.04},
				{X: 0.8, Y: 0.08},
				{X: 1, Y: 0},
				{X: 0.8, Y: -0.08},
				{X: 0.8, Y: -0.04},
			}
			var d string
			plen := len(unitPoints)
			for i, point := range unitPoints {
				orix := point.X
				oriy := point.Y
				fromx := annotation.FromPoint.X * float32(vstream.Width)
				fromy := annotation.FromPoint.Y * float32(vstream.Height)
				tox := annotation.ToPoint.X * float32(vstream.Width)
				toy := annotation.ToPoint.Y * float32(vstream.Height)
				// 根据变换矩阵，变换后的点坐标(A, B)为
				// A = a(x2 - x1) - b(y2 - y1) + x1
				// B = a(y2 - y1) + b(x2 - x1) + y1
				x := int(orix*(tox-fromx) - oriy*(toy-fromy) + fromx)
				y := int(orix*(toy-fromy) + oriy*(tox-fromx) + fromy)
				if i == 0 {
					d = fmt.Sprintf("%sM%d %d ", d, x, y)
				} else if i == plen-1 {
					d = fmt.Sprintf("%sL%d %d Z", d, x, y)
				} else {
					d = fmt.Sprintf("%sL%d %d ", d, x, y)
				}
			}
			var styles []string
			if annotation.Stroke != "" {
				styles = append(styles, fmt.Sprintf("stroke:%s", annotation.Stroke), fmt.Sprintf("fill:%s", annotation.Stroke))
			}
			if annotation.StrokeWidth != 0 {
				styles = append(styles, fmt.Sprintf("stroke-width:%dpx", annotation.StrokeWidth))
			}
			canvas.Path(d, strings.Join(styles, ";"))

		case "text":
			fromx := int(annotation.FromPoint.X * float32(vstream.Width))
			fromy := int(annotation.FromPoint.Y * float32(vstream.Height))
			var styles []string
			if annotation.Stroke != "" {
				styles = append(styles, fmt.Sprintf("fill:%s", annotation.Stroke))
			}
			if annotation.FontSize != 0 {
				styles = append(styles, fmt.Sprintf("font-size:%dpx", annotation.FontSize))
			}
			canvas.Text(fromx, fromy, annotation.Text, strings.Join(styles, ";"))
		}
	}
	canvas.End()

	cc := exec.CommandContext(context.Background(), "rsvg-convert", svglocalfile, "-f", "png", "-o", params.Outfile)
	_, err = cc.CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}
