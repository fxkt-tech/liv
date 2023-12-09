package liv

import (
	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffprobe"
)

type Option func(*options)

type options struct {
	ffmpegOpts  []ffmpeg.Option
	ffprobeOpts []ffprobe.Option
}

func FFmpegOptions(ffmpegOpts ...ffmpeg.Option) Option {
	return func(o *options) {
		o.ffmpegOpts = ffmpegOpts
	}
}

func FFprobeOptions(ffprobeOpts ...ffprobe.Option) Option {
	return func(o *options) {
		o.ffprobeOpts = ffprobeOpts
	}
}
