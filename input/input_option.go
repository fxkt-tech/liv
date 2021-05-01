package input

type OptionFunc func(*option)

type option struct {
	i   string   // i is input file.
	ss  int64    // ss is starttime.
	t   int64    // t is duration.
	ext []string // extra params.
}

func SetI(i string) OptionFunc {
	return func(o *option) {
		o.i = i
	}
}
