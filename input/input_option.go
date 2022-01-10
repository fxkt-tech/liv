package input

type OptionFunc func(*option)

type option struct {
	i   string   // i is input file.
	ss  float64  // ss is start_time.
	t   float64  // t is duration.
	ext []string // extra params.
}

func SetI(i string) OptionFunc {
	return func(o *option) {
		o.i = i
	}
}

func SetStartTime(ss float64) OptionFunc {
	return func(o *option) {
		o.ss = ss
	}
}

func SetDuration(t float64) OptionFunc {
	return func(o *option) {
		o.t = t
	}
}
