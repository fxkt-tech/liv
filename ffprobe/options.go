package ffprobe

type Option func(*FFprobe)

func WithBin(bin string) Option {
	return func(fp *FFprobe) {
		fp.bin = bin
	}
}

func WithDebug(debug bool) Option {
	return func(fp *FFprobe) {
		fp.debug = debug
	}
}
