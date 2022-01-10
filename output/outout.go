package output

import (
	"fmt"
	"strconv"
	"strings"
)

// Output is common output info.
type Output struct {
	opt *option
}

func New(opts ...OptionFunc) *Output {
	o := &option{
		// threads:               4,
		// max_muxing_queue_size: 4086,
		// movflags: "faststart",
		// cv:                    "libx264",
		// ca:                    "copy",
		// hls_time: 2,
	}
	for _, opt := range opts {
		opt(o)
	}
	return &Output{
		opt: o,
	}
}

func (o *Output) Params() (params []string) {
	if len(o.opt.maps) != 0 {
		for _, m := range o.opt.maps {
			if strings.Contains(m, ":") {
				params = append(params, "-map", fmt.Sprintf("%s", m))
			} else {
				params = append(params, "-map", fmt.Sprintf("[%s]", m))
			}
		}
	}
	if o.opt.cv != "" {
		params = append(params, "-c:v", o.opt.cv)
	}
	if o.opt.ca != "" {
		params = append(params, "-c:a", o.opt.ca)
	}
	if o.opt.bv != 0 {
		params = append(params, "-b:v", strconv.FormatInt(int64(o.opt.bv), 10))
	}
	if o.opt.ba != 0 {
		params = append(params, "-b:a", strconv.FormatInt(int64(o.opt.ba), 10))
	}
	if o.opt.vframes != 0 {
		params = append(params, "-vframes", strconv.FormatInt(int64(o.opt.vframes), 10))
	}
	if len(o.opt.metadatas) != 0 {
		for _, m := range o.opt.metadatas {
			params = append(params, "-metadata", m)
		}
	}
	if o.opt.f == "hls" {
		if o.opt.hls_segment_type != "" {
			params = append(params, "-hls_segment_type", o.opt.hls_segment_type)
		}
		if o.opt.hls_flags != "" {
			params = append(params, "-hls_flags", o.opt.hls_flags)
		}
		if o.opt.hls_playlist_type != "" {
			params = append(params, "-hls_playlist_type", o.opt.hls_playlist_type)
		}
		if o.opt.hls_time > 0 {
			params = append(params, "-hls_time", strconv.FormatInt(int64(o.opt.hls_time), 10))
			params = append(params, "-g", strconv.FormatInt(int64(o.opt.hls_time), 10))
		}
		if o.opt.master_pl_name != "" {
			params = append(params, "-master_pl_name", o.opt.master_pl_name)
		}
		if o.opt.hls_segment_filename != "" {
			params = append(params, "-hls_segment_filename", o.opt.hls_segment_filename)
		}
		params = append(params, "-hls_list_size", "0")
	}
	if o.opt.threads != 0 {
		params = append(params, "-threads", strconv.FormatInt(int64(o.opt.threads), 10))
	}
	if o.opt.max_muxing_queue_size != 0 {
		params = append(params, "-max_muxing_queue_size", strconv.FormatInt(int64(o.opt.max_muxing_queue_size), 10))
	}
	if o.opt.movflags != "" {
		params = append(params, "-movflags", o.opt.movflags)
	}
	if o.opt.ss != 0 {
		params = append(params, "-ss", strconv.FormatFloat(o.opt.ss, 'f', -1, 64))
	}
	if o.opt.t != 0 {
		params = append(params, "-t", strconv.FormatFloat(o.opt.t, 'f', -1, 64))
	}
	if o.opt.f != "" {
		params = append(params, "-f", o.opt.f)
	}
	if o.opt.file != "" {
		params = append(params, o.opt.file)
	}
	return
}

func (o *Output) String() string {
	return strings.Join(o.Params(), " ")
}
