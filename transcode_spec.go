package liv

type TranscodeSpec struct{}

func NewTranscodeSpec() *TranscodeSpec {
	return &TranscodeSpec{}
}

func (*TranscodeSpec) SimpleMP4Satified(params *TranscodeParams) error {
	if params == nil || len(params.Subs) == 0 {
		return ErrParamsInvalid
	}

	return nil
}

func (*TranscodeSpec) SimpleMP3Satified(params *TranscodeParams) error {
	if params == nil || len(params.Subs) == 0 {
		return ErrParamsInvalid
	}

	return nil
}

func (*TranscodeSpec) SimpleJPEGSatified(params *TranscodeParams) error {
	if params == nil || len(params.Subs) == 0 {
		return ErrParamsInvalid
	}

	return nil
}

func (*TranscodeSpec) ConvertContainerSatified(params *ConvertContainerParams) error {
	if params == nil || params.InFile == "" || params.OutFile == "" {
		return ErrParamsInvalid
	}
	return nil
}
