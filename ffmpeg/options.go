package ffmpeg

type Option func(*FFmpeg)

func WithBin(bin string) Option {
	return func(fm *FFmpeg) {
		fm.bin = bin
	}
}

func WithDebug(debug bool) Option {
	return func(fm *FFmpeg) {
		fm.debug = debug
	}
}

func WithOverwrite(y bool) Option {
	return func(fm *FFmpeg) {
		fm.y = y
	}
}

// log level
func WithLogLevel(v LogLevel) Option {
	return func(fm *FFmpeg) {
		fm.v = v
	}
}

func WithDry(dry bool) Option {
	return func(f *FFmpeg) {
		f.dry = dry
	}
}
