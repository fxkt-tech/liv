package output

import (
	"strconv"
	"strings"

	"github.com/fxkt-tech/liv/ffmpeg/codec"
	"github.com/fxkt-tech/liv/ffmpeg/stream"
)

// Output is common output info.
type Output struct {
	maps                  []stream.Streamer // mean is -map.
	cv, ca                string            // cv is c:v, ca is c:a.
	shortest              bool              // -shortest
	bv, ba                int32             // bv is b:v, ba is b:a.
	pix_fmt               string
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
	vsync                 string
	g                     int32 // gop

	nonSupportArgs map[string]string

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

func New(opts ...Option) *Output {
	op := &Output{
		// threads:               4,
		// max_muxing_queue_size: 4086,
		// movflags: "faststart",
		// cv:                    codec.X264,
		// ca:                    codec.Copy,
		// hls_time:              2,
		nonSupportArgs: make(map[string]string),
	}
	for _, o := range opts {
		o(op)
	}
	return op
}

func (o *Output) Params() (params []string) {
	if len(o.maps) != 0 {
		for _, m := range o.maps {
			params = append(params, "-map", m.Name())
		}
	}
	if o.cv != "" {
		if o.cv != codec.Nope {
			params = append(params, "-c:v", o.cv)
		} else {
			params = append(params, "-vn")
		}
	}
	if o.ca != "" {
		if o.ca != codec.Nope {
			params = append(params, "-c:a", o.ca)
		} else {
			params = append(params, "-an")
		}
	}
	if o.shortest {
		params = append(params, "-shortest")
	}
	if o.bv != 0 {
		params = append(params, "-b:v", strconv.FormatInt(int64(o.bv), 10))
	}
	if o.ba != 0 {
		params = append(params, "-b:a", strconv.FormatInt(int64(o.ba), 10))
	}
	if o.pix_fmt != "" {
		params = append(params, "-pix_fmt", o.pix_fmt)
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
	for k, v := range o.nonSupportArgs {
		params = append(params, "-"+k, v)
	}
	if o.var_stream_map != "" {
		params = append(params, "-var_stream_map", o.var_stream_map)
	}
	if o.vsync != "" {
		params = append(params, "-vsync", o.vsync)
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
			if o.g == 0 {
				o.g = o.hls_time
			}
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
	if o.g > 0 {
		params = append(params, "-g", strconv.FormatInt(int64(o.g), 10))
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
