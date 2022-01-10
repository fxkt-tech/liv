package ffmpeg

type OptionFunc func(*option)

type option struct {
	nopFilterComplex bool // 不需要复合滤镜
}

func NopFilterComplex(b bool) OptionFunc {
	return func(o *option) {
		o.nopFilterComplex = b
	}
}
