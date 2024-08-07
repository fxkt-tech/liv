package ffprobe

type Probe struct {
	Streams []*ProbeStream `json:"streams,omitempty"`
	Format  *ProbeFormat   `json:"format,omitempty"`
}

type ProbeStream struct {
	// common
	Index     int32   `json:"index,omitempty"`
	CodecType string  `json:"codec_type,omitempty"`
	CodecName string  `json:"codec_name,omitempty"`
	Profile   string  `json:"profile,omitempty"`
	BitRate   int32   `json:"bit_rate,omitempty,string"`
	Duration  float32 `json:"duration,omitempty,string"`

	// video
	Width      int32  `json:"width,omitempty"`
	Height     int32  `json:"height,omitempty"`
	HasBFrames int32  `json:"has_b_frames,omitempty"`
	PixFmt     string `json:"pix_fmt,omitempty"`
	Level      int32  `json:"level,omitempty"`
	RFrameRate string `json:"r_frame_rate,omitempty"`
	SAR        string `json:"sample_aspect_radio,omitempty"`
	DAR        string `json:"display_aspect_radio,omitempty"`
	NBFrames   int32  `json:"nb_frames,omitempty,string"`

	// audio
	SampleFmt     string `json:"sample_fmt,omitempty"`         // 采样格式
	SampleRate    int32  `json:"sample_rate,omitempty,string"` // 采样率
	Channels      int32  `json:"channels,omitempty"`           // 声道数
	ChannelLayout string `json:"channel_layout,omitempty"`     // 声道布局
}

type ProbeFormat struct {
	FormatName string  `json:"format_name,omitempty"`
	Size       int64   `json:"size,omitempty,string"`
	Duration   float32 `json:"duration,omitempty,string"`
}
