package output

import (
	"fmt"

	"github.com/fxkt-tech/liv/ffmpeg/stream"
)

type Option func(*Output)

func Map(m stream.Streamer) Option {
	return func(o *Output) {
		o.maps = append(o.maps, m)
	}
}

func MovFlags(flag string) Option {
	return func(o *Output) {
		o.movflags = flag
	}
}

func Thread(n int32) Option {
	return func(o *Output) {
		if n > 0 {
			o.threads = n
		}
	}
}

func MaxMuxingQueueSize(sz int32) Option {
	return func(o *Output) {
		o.max_muxing_queue_size = sz
	}
}

func VideoCodec(cv string) Option {
	return func(o *Output) {
		o.cv = cv
	}
}

func AudioCodec(ca string) Option {
	return func(o *Output) {
		o.ca = ca
	}
}

func VideoBitrate(bv int32) Option {
	return func(o *Output) {
		o.bv = bv
	}
}

func PixFmt(pixFmt string) Option {
	return func(o *Output) {
		o.pix_fmt = pixFmt
	}
}

func Crf(crf int32) Option {
	return func(o *Output) {
		o.crf = crf
	}
}

func AudioBitrate(ba int32) Option {
	return func(o *Output) {
		o.ba = ba
	}
}

func Metadata(k, v string) Option {
	return func(o *Output) {
		o.metadatas = append(o.metadatas, fmt.Sprintf("%s=%s", k, v))
	}
}

func StartTime(ss float64) Option {
	return func(o *Output) {
		o.ss = ss
	}
}

func Duration(t float64) Option {
	return func(o *Output) {
		o.t = t
	}
}

func File(f string) Option {
	return func(o *Output) {
		o.file = f
	}
}

func Format(f string) Option {
	return func(o *Output) {
		o.f = f
	}
}

func VarStreamMap(s string) Option {
	return func(o *Output) {
		o.var_stream_map = s
	}
}

func VSync(vsync string) Option {
	return func(o *Output) {
		o.vsync = vsync
	}
}

func GOP(g int32) Option {
	return func(o *Output) {
		o.g = g
	}
}

// hls

func HLSSegmentType(value string) Option {
	return func(o *Output) {
		o.hls_segment_type = value
	}
}

func HLSFlags(value string) Option {
	return func(o *Output) {
		o.hls_flags = value
	}
}

func HLSPlaylistType(value string) Option {
	return func(o *Output) {
		o.hls_playlist_type = value
	}
}

func HLSTime(value int32) Option {
	return func(o *Output) {
		o.hls_time = value
	}
}

func MasterPlName(value string) Option {
	return func(o *Output) {
		o.master_pl_name = value
	}
}

func HLSSegmentFilename(value string) Option {
	return func(o *Output) {
		o.hls_segment_filename = value
	}
}

func HLSKeyInfoFile(f string) Option {
	return func(o *Output) {
		o.hls_key_info_file = f
	}
}

// img

func Vframes(vframes int32) Option {
	return func(o *Output) {
		o.vframes = vframes
	}
}

func KV(k, v string) Option {
	return func(o *Output) {
		o.nonSupportArgs[k] = v
	}
}
