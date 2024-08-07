package liv

type ConvertContainerParams struct {
	InFile   string
	OutFile  string
	Metadata []*KV
	Threads  int32
}

type TranscodeParams struct {
	Infile string
	Subs   []*SubTranscodeParams
}
type SubTranscodeParams struct {
	Outfile string
	Filters *Filters
	Threads int32
}

type TranscodeSimpleTSParams struct {
	Infile  string
	Outfile string
	Filters *Filters
	Threads int32
}

type TranscodeSimpleHLSParams struct {
	Infile  string
	Outfile string
	Filters *Filters
	Threads int32
}

type ConcatParams struct {
	Infiles    []string
	ConcatFile string // eg. mylist.txt
	Outfile    string
	Duration   float32
}

type ExtractAudioParams struct {
	Infile  string
	Outfile string
}

type MergeParams struct {
	FramesInfile string
	VideoInfile  string
	AudioInfile  string
	Filters      *Filters
	Outfile      string
}

type Filters struct {
	Container string    `json:"container,omitempty"`
	Metadata  []*KV     `json:"metadata,omitempty"`
	Video     *Video    `json:"video,omitempty"`
	Audio     *Audio    `json:"audio,omitempty"`
	Logo      []*Logo   `json:"logo,omitempty"`
	Delogo    []*Delogo `json:"delogo,omitempty"`
	Clip      *Clip     `json:"clip,omitempty"`
	HLS       *HLS      `json:"hls,omitempty"`
}

type KV struct {
	K string `json:"k,omitempty"`
	V string `json:"v,omitempty"`
}

type Video struct {
	Codec     string  `json:"codec,omitempty"`
	Width     int32   `json:"width,omitempty"`
	Height    int32   `json:"height,omitempty"`
	Short     int32   `json:"short,omitempty"`
	FPS       string  `json:"fps,omitempty"`
	Crf       int32   `json:"crf,omitempty"`
	WZQuality float32 `json:"wz_quality,omitempty"`
	Bitrate   int32   `json:"bitrate,omitempty"`
	GOP       int32   `json:"gop,omitempty"`
	PTS       string  `json:"pts,omitempty"`
	APTS      string  `json:"apts,omitempty"`
	PixFmt    string  `json:"pix_fmt,omitempty"`
}

type Audio struct {
	Codec   string `json:"codec,omitempty"`
	Bitrate int32  `json:"bitrate,omitempty"`
}

type Logo struct {
	File string  `json:"file,omitempty"`
	Pos  string  `json:"pos,omitempty"`
	Dx   float32 `json:"dx,omitempty"`
	Dy   float32 `json:"dy,omitempty"`
	LW   float32 `json:"lw,omitempty"`
	LH   float32 `json:"lh,omitempty"`
}

func (l *Logo) NeedScale() bool {
	return l.LW > 0 || l.LH > 0
}

// 矩形框
type Rectangle struct {
	X float32 `json:"x,omitempty"`
	Y float32 `json:"y,omitempty"`
	W float32 `json:"w,omitempty"`
	H float32 `json:"h,omitempty"`
}

type Delogo struct {
	SS   float32    `json:"ss,omitempty"`
	Rect *Rectangle `json:"rect,omitempty"`
}

type Clip struct {
	Seek     float32 `json:"seek,omitempty"`
	Duration float32 `json:"duration,omitempty"`
}

type HLS struct {
	HLSSegmentType     string `json:"hls_segment_type,omitempty"`
	HLSFlags           string `json:"hls_flags,omitempty"`
	HLSPlaylistType    string `json:"hls_playlist_type,omitempty"`
	HLSTime            int32  `json:"hls_time,omitempty"`
	MasterPlName       string `json:"master_pl_name,omitempty"`
	HLSSegmentFilename string `json:"hls_segment_filename,omitempty"`
}
