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

func (*TranscodeSpec) SimpleHLSSatified(params *TranscodeSimpleHLSParams) error {
	if params == nil || params.Infile == "" || params.Outfile == "" {
		return ErrParamsInvalid
	}
	return nil
}

func (*TranscodeSpec) SimpleTSSatified(params *TranscodeSimpleTSParams) error {
	if params == nil || params.Infile == "" || params.Outfile == "" {
		return ErrParamsInvalid
	}

	return nil
}

func (*TranscodeSpec) ExtractAudioSatified(params *ExtractAudioParams) error {
	if params == nil || params.Infile == "" || params.Outfile == "" {
		return ErrParamsInvalid
	}

	return nil
}

func (*TranscodeSpec) MergeByFramesSatified(params *MergeParams) error {
	if params == nil ||
		params.FramesInfile == "" ||
		params.AudioInfile == "" ||
		params.Outfile == "" {
		return ErrParamsInvalid
	}
	return nil
}
