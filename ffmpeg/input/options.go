package input

type Option func(*Input)

func I(i string) Option {
	return func(o *Input) {
		o.i = i
	}
}

func VideoCodec(cv string) Option {
	return func(o *Input) {
		o.cv = cv
	}
}

func StartTime(ss float64) Option {
	return func(o *Input) {
		o.ss = ss
	}
}

func Duration(t float64) Option {
	return func(o *Input) {
		o.t = t
	}
}

func FPS(fps string) Option {
	return func(i *Input) {
		i.r = fps
	}
}

func Format(f string) Option {
	return func(i *Input) {
		i.f = f
	}
}

func Safe(safe string) Option {
	return func(i *Input) {
		i.safe = safe
	}
}
