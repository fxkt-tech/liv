package input

type Inputs []*Input

func (inputs Inputs) Params() (params []string) {
	for _, input := range inputs {
		params = append(params, input.Params()...)
	}
	return
}
