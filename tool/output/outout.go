package output

import (
	"fmt"
	"strconv"
	"strings"

	"fxkt.tech/echo/tool/codec"
)

type OutputOption func(*Output)

func Map(m string) OutputOption {
	return func(o *Output) {
		o.maps = append(o.maps, m)
	}
}

func MovFlags(flag string) OutputOption {
	return func(o *Output) {
		o.movflags = flag
	}
}

func Thread(n int32) OutputOption {
	return func(o *Output) {
		if n > 0 {
			o.threads = n
		}
	}
}

func MaxMuxingQueueSize(sz int32) OutputOption {
	return func(o *Output) {
		o.max_muxing_queue_size = sz
	}
}

func VideoCodec(cv string) OutputOption {
	return func(o *Output) {
		o.cv = cv
	}
}

func AudioCodec(ca string) OutputOption {
	return func(o *Output) {
		o.ca = ca
	}
}

func VideoBitrate(bv int32) OutputOption {
	return func(o *Output) {
		o.bv = bv
	}
}

func Crf(crf int32) OutputOption {
	return func(o *Output) {
		o.crf = crf
	}
}

func AudioBitrate(ba int32) OutputOption {
	return func(o *Output) {
		o.ba = ba
	}
}

func Metadata(k, v string) OutputOption {
	return func(o *Output) {
		o.metadatas = append(o.metadatas, fmt.Sprintf("%s=%s", k, v))
	}
}

func StartTime(ss float64) OutputOption {
	return func(o *Output) {
		o.ss = ss
	}
}

func Duration(t float64) OutputOption {
	return func(o *Output) {
		o.t = t
	}
}

func File(f string) OutputOption {
	return func(o *Output) {
		o.file = f
	}
}

func Format(f string) OutputOption {
	return func(o *Output) {
		o.f = f
	}
}

func VarStreamMap(s string) OutputOption {
	return func(o *Output) {
		o.var_stream_map = s
	}
}

// hls

func HlsSegmentType(value string) OutputOption {
	return func(o *Output) {
		o.hls_segment_type = value
	}
}

func HlsFlags(value string) OutputOption {
	return func(o *Output) {
		o.hls_flags = value
	}
}

func HlsPlaylistType(value string) OutputOption {
	return func(o *Output) {
		o.hls_playlist_type = value
	}
}
func HlsTime(value int32) OutputOption {
	return func(o *Output) {
		o.hls_time = value
	}
}
func MasterPlName(value string) OutputOption {
	return func(o *Output) {
		o.master_pl_name = value
	}
}
func HlsSegmentFilename(value string) OutputOption {
	return func(o *Output) {
		o.hls_segment_filename = value
	}
}

func HlsKeyInfoFile(f string) OutputOption {
	return func(o *Output) {
		o.hls_key_info_file = f
	}
}

// img

func Vframes(vframes int32) OutputOption {
	return func(o *Output) {
		o.vframes = vframes
	}
}

// Output is common output info.
type Output struct {
	maps                  []string // mean is -map.
	cv, ca                string   // cv is c:v, ca is c:a.
	bv, ba                int32    // bv is b:v, ba is b:a.
	crf                   int32
	metadatas             []string // mean is -metadata.
	threads               int32    // thread counts, default 4.
	max_muxing_queue_size int32    // queue size when muxing, default 4086.
	movflags              string   // location of mp4 moov.
	ss                    float64  // ss is start_time.
	t                     float64  // t is duration.
	f                     string   // f is -f format.
	file                  string
	var_stream_map        string

	// hls configs
	hls_segment_type     string
	hls_flags            string
	hls_playlist_type    string
	hls_time             int32
	master_pl_name       string
	hls_segment_filename string
	hls_key_info_file    string // 加密

	// img
	vframes int32
}

func New(opts ...OutputOption) *Output {
	op := &Output{
		threads:               4,
		max_muxing_queue_size: 4086,
		movflags:              "faststart",
		cv:                    codec.X264,
		ca:                    codec.Copy,
		hls_time:              2,
	}
	for _, o := range opts {
		o(op)
	}
	return op
}

func (o *Output) Params() (params []string) {
	if len(o.maps) != 0 {
		for _, m := range o.maps {
			params = append(params, "-map", m)
		}
	}
	if o.cv != "" {
		params = append(params, "-c:v", o.cv)
	}
	if o.ca != "" {
		params = append(params, "-c:a", o.ca)
	}
	if o.bv != 0 {
		params = append(params, "-b:v", strconv.FormatInt(int64(o.bv), 10))
	}
	if o.ba != 0 {
		params = append(params, "-b:a", strconv.FormatInt(int64(o.ba), 10))
	}
	if o.crf != 0 {
		params = append(params, "-crf", strconv.FormatInt(int64(o.crf), 10))
	}
	if o.vframes != 0 {
		params = append(params, "-vframes", strconv.FormatInt(int64(o.vframes), 10))
	}
	if len(o.metadatas) != 0 {
		for _, m := range o.metadatas {
			params = append(params, "-metadata", m)
		}
	}
	if o.var_stream_map != "" {
		params = append(params, "-var_stream_map", o.var_stream_map)
	}
	if o.f == "hls" {
		if o.hls_segment_type != "" {
			params = append(params, "-hls_segment_type", o.hls_segment_type)
		}
		if o.hls_flags != "" {
			params = append(params, "-hls_flags", o.hls_flags)
		}
		if o.hls_playlist_type != "" {
			params = append(params, "-hls_playlist_type", o.hls_playlist_type)
		}
		if o.hls_time > 0 {
			params = append(params, "-hls_time", strconv.FormatInt(int64(o.hls_time), 10))
			params = append(params, "-g", strconv.FormatInt(int64(o.hls_time), 10))
		}
		if o.master_pl_name != "" {
			params = append(params, "-master_pl_name", o.master_pl_name)
		}
		if o.hls_segment_filename != "" {
			params = append(params, "-hls_segment_filename", o.hls_segment_filename)
		}
		if o.hls_key_info_file != "" {
			params = append(params, "-hls_key_info_file", o.hls_key_info_file)
		}
		params = append(params, "-hls_list_size", "0")
	}
	if o.threads != 0 {
		params = append(params, "-threads", strconv.FormatInt(int64(o.threads), 10))
	}
	if o.max_muxing_queue_size != 0 {
		params = append(params, "-max_muxing_queue_size", strconv.FormatInt(int64(o.max_muxing_queue_size), 10))
	}
	if o.movflags != "" {
		params = append(params, "-movflags", o.movflags)
	}
	if o.ss != 0 {
		params = append(params, "-ss", strconv.FormatFloat(o.ss, 'f', -1, 64))
	}
	if o.t != 0 {
		params = append(params, "-t", strconv.FormatFloat(o.t, 'f', -1, 64))
	}
	if o.f != "" {
		params = append(params, "-f", o.f)
	}
	if o.file != "" {
		params = append(params, o.file)
	}
	return
}

func (o *Output) String() string {
	return strings.Join(o.Params(), " ")
}

// outputs

type Outputs []*Output

func (outputs Outputs) Params() (params []string) {
	for _, output := range outputs {
		params = append(params, output.Params()...)
	}
	return
}

func (outputs Outputs) String() string {
	return strings.Join(outputs.Params(), " ")
}
