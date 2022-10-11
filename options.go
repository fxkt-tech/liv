package liv

import "github.com/fxkt-tech/liv/ffmpeg"

type Option func(*options)

type options struct {
	ffmpegOpts  []ffmpeg.FFmpegOption
	ffprobeOpts []ffmpeg.FFprobeOption
}

func FFmpegOptions(ffmpegOpts ...ffmpeg.FFmpegOption) Option {
	return func(o *options) {
		o.ffmpegOpts = ffmpegOpts
	}
}

func FFprobeOptions(ffprobeOpts ...ffmpeg.FFprobeOption) Option {
	return func(o *options) {
		o.ffprobeOpts = ffprobeOpts
	}
}
