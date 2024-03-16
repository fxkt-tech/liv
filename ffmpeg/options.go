package ffmpeg

type Option func(*FFmpeg)

func Binary(bin string) Option {
	return func(fm *FFmpeg) {
		fm.bin = bin
	}
}

func Debug(debug bool) Option {
	return func(fm *FFmpeg) {
		fm.debug = debug
	}
}

func Overwrite(y bool) Option {
	return func(fm *FFmpeg) {
		fm.y = y
	}
}

// log level
func V(v LogLevel) Option {
	return func(fm *FFmpeg) {
		fm.v = v
	}
}

func Dry(dry bool) Option {
	return func(f *FFmpeg) {
		f.dry = dry
	}
}
