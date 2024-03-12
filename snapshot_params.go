package liv

type SnapshotParams struct {
	Infile    string
	Outfile   string
	StartTime float64
	Interval  int32
	Num       int32
	FrameType int32 // 0-仅关键帧 1-指定时间点的帧
	Width     int32
	Height    int32
}

type SpriteParams struct {
	Infile   string
	Outfile  string
	XLen     int32
	YLen     int32
	Width    int32
	Height   int32
	Interval float64
}

type SVGMarkParams struct {
	Infile      string
	Outfile     string
	StartTime   float64          `json:"start_time,omitempty"`
	Annotations []*SVGAnnotation `json:"annotation,omitempty"`
}

type SVGAnnotation struct {
	Type        string   `json:"type,omitempty"`
	Stroke      string   `json:"color,omitempty"`
	Text        string   `json:"text,omitempty"`
	Points      []*Point `json:"points,omitempty"`
	FromPoint   *Point   `json:"from_point,omitempty"`
	ToPoint     *Point   `json:"to_point,omitempty"`
	StrokeWidth int32    `json:"stroke_width,omitempty"`
	FontSize    int32    `json:"font_size,omitempty"`
}

type Point struct {
	X float64 `json:"x,omitempty"`
	Y float64 `json:"y,omitempty"`
}
