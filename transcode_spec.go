package liv

type TranscodeSpec struct{}

func NewTranscodeSpec() *TranscodeSpec {
	return &TranscodeSpec{}
}

func (*TranscodeSpec) CheckSatified(params *TranscodeParams) error {
	if params == nil || len(params.Subs) == 0 {
		return ErrParamsInvalid
	}
	return nil
}
