package liv

type SnapshotParams struct {
	Infile  string
	Outfile string

	// 0-仅关键帧截图 1-等间隔截图 2-指定帧序列截图
	FrameType int32 // 截图类型

	// 等间隔截图
	Interval       int32 // 间隔时间
	IntervalFrames int32 // 间隔帧数

	// 指定帧序列截图
	Frames []int32 // 帧序列

	StartTime float32 // 截图开始时间。截图类型为指定帧序列截图时不建议使用此参数
	Num       int32   // 截图最大数量
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
	Interval float32
}

type SVGMarkParams struct {
	Infile      string
	Outfile     string
	StartTime   float32          `json:"start_time,omitempty"`
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
	X float32 `json:"x,omitempty"`
	Y float32 `json:"y,omitempty"`
}
