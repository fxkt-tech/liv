package fusion

import "github.com/fxkt-tech/liv/ffmpeg"

type ShelfOption func(*TrackData)

func WithStageSize(w, h int32) ShelfOption {
	return func(td *TrackData) {
		td.stageWidth, td.stageHeight = w, h
	}
}

func WithFFmpegOptions(opts ...ffmpeg.Option) ShelfOption {
	return func(td *TrackData) {
		td.ffmpegOpts = append(td.ffmpegOpts, opts...)
	}
}
