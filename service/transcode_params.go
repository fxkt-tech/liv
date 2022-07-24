package echosrv

type TranscodeParams struct {
	Infile string
	Subs   []*SubTranscodeParams
}
type SubTranscodeParams struct {
	Outfile string
	Filters *TranscodeFilters
}

type TranscodeFilters struct {
	Container string             `json:"container,omitempty"`
	Video     *TranscodeVideo    `json:"video,omitempty"`
	Audio     *TranscodeAudio    `json:"audio,omitempty"`
	Logo      []*TranscodeLogo   `json:"logo,omitempty"`
	Delogo    []*TranscodeDelogo `json:"delogo,omitempty"`
	Clip      *TranscodeClip     `json:"clip,omitempty"`
	Hls       *TranscodeHls      `json:"hls,omitempty"`
}

type TranscodeVideo struct {
	Width     int32   `json:"width,omitempty"`
	Height    int32   `json:"height,omitempty"`
	Crf       int32   `json:"crf,omitempty"`
	WZQuality float32 `json:"wz_quality,omitempty"`
	Bitrate   int32   `json:"bitrate,omitempty"`
}

type TranscodeAudio struct {
	Bitrate int32 `json:"bitrate,omitempty"`
}

type TranscodeLogo struct {
	File string  `json:"file,omitempty"`
	Pos  string  `json:"pos,omitempty"`
	Dx   float64 `json:"dx,omitempty"`
	Dy   float64 `json:"dy,omitempty"`
	LW   float64 `json:"lw,omitempty"`
	LH   float64 `json:"lh,omitempty"`
}

// 矩形框
type Rectangle struct {
	X float64 `json:"x,omitempty"`
	Y float64 `json:"y,omitempty"`
	W float64 `json:"w,omitempty"`
	H float64 `json:"h,omitempty"`
}

type TranscodeDelogo struct {
	SS    float64      `json:"ss,omitempty"`
	Rects []*Rectangle `json:"rects,omitempty"`
}

type TranscodeClip struct {
	Seek     float64 `json:"seek,omitempty"`
	Duration float64 `json:"duration,omitempty"`
}

type TranscodeHls struct {
	HlsSegmentType     string `json:"hls_segment_type,omitempty"`
	HlsFlags           string `json:"hls_flags,omitempty"`
	HlsPlaylistType    string `json:"hls_playlist_type,omitempty"`
	HlsTime            int32  `json:"hls_time,omitempty"`
	MasterPlName       string `json:"master_pl_name,omitempty"`
	HlsSegmentFilename string `json:"hls_segment_filename,omitempty"`
}
