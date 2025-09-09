package fusion

import "github.com/fxkt-tech/liv/ffmpeg"

type Template struct {
	trackData *TrackData
}

func NewTemplate(opts ...ShelfOption) *Template {
	trackData := NewTrackData(
		WithStageSize(960, 540),
		WithFFmpegOptions(
			ffmpeg.WithDebug(true),
		),
	)

	d := &Template{trackData: trackData}

	return d
}

func (t *Template) FromProto(proto string) error {
	return nil
}
