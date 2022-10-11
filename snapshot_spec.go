package liv

type SnapshotSpec struct{}

func NewSnapshotSpec() *SnapshotSpec {
	return &SnapshotSpec{}
}

func (*SnapshotSpec) CheckSatified(params *SnapshotParams) error {
	if params == nil {
		return ErrParamsInvalid
	}
	if params.FrameType == 1 && params.Num > 1 && params.Interval <= 0 {
		return ErrParamsInvalid2
	}
	return nil
}
