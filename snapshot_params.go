package liv

type SnapshotParams struct {
	Infile    string
	Outfile   string
	StartTime float64
	Interval  int32
	Num       int32
	FrameType int32
	NotBlack  bool
	NotWhite  bool
}
