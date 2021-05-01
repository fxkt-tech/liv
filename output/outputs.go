package output

type Outputs []*Output

func (Outputs Outputs) Params() (params []string) {
	for _, output := range Outputs {
		params = append(params, output.Params()...)
	}
	return
}
