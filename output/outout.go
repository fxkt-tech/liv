package output

import (
	"fmt"
	"strconv"
	"strings"
)

// Output is common output info.
type Output struct {
	// maps                  []string // mean is -map.
	// cv, ca                string   // cv is c:v, ca is c:a.
	// metadatas             []string // mean is -metadata.
	// threads               int32    // thread counts, default 4.
	// max_muxing_queue_size int32    // queue size when muxing, default 4086.
	// movflags              string   // location of mp4 moov.
	// f                     string   // f is -f format.
	// file                  string
	opt *option
}

func New(opts ...OptionFunc) *Output {
	o := &option{
		threads:               4,
		max_muxing_queue_size: 4086,
		movflags:              "faststart",
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
			params = append(params, "-map", fmt.Sprintf("[%s]", m))
		}
	}
	if o.opt.cv != "" {
		params = append(params, "c:v", o.opt.cv)
	}
	if o.opt.ca != "" {
		params = append(params, "c:a", o.opt.ca)
	}
	if len(o.opt.metadatas) != 0 {
		for _, m := range o.opt.metadatas {
			params = append(params, "-metadata", m)
		}
	}
	if o.opt.max_muxing_queue_size != 0 {
		params = append(params, "-max_muxing_queue_size", strconv.FormatInt(int64(o.opt.max_muxing_queue_size), 10))
	}
	if o.opt.movflags != "" {
		params = append(params, "-movflags", o.opt.movflags)
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
