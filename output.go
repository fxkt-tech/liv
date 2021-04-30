package ffmpeg

type Output struct {
	params []string
}

func NewOutput(params ...string) *Output {
	return &Output{params: params}
}

func (o *Output) AddParams(params ...string) {
	o.params = append(o.params, params...)
}
