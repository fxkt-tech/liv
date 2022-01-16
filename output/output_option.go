package output

import "fmt"

type OptionFunc func(*option)

type option struct {
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

func Map(m string) OptionFunc {
	return func(o *option) {
		o.maps = append(o.maps, m)
	}
}

func MovFlags(flag string) OptionFunc {
	return func(o *option) {
		o.movflags = flag
	}
}

func Thread(n int32) OptionFunc {
	return func(o *option) {
		if n > 0 {
			o.threads = n
		}
	}
}

func MaxMuxingQueueSize(sz int32) OptionFunc {
	return func(o *option) {
		o.max_muxing_queue_size = sz
	}
}

func VideoCodec(cv string) OptionFunc {
	return func(o *option) {
		o.cv = cv
	}
}

func AudioCodec(ca string) OptionFunc {
	return func(o *option) {
		o.ca = ca
	}
}

func VideoBitrate(bv int32) OptionFunc {
	return func(o *option) {
		o.bv = bv
	}
}

func Crf(crf int32) OptionFunc {
	return func(o *option) {
		o.crf = crf
	}
}

func AudioBitrate(ba int32) OptionFunc {
	return func(o *option) {
		o.ba = ba
	}
}

func Metadata(k, v string) OptionFunc {
	return func(o *option) {
		o.metadatas = append(o.metadatas, fmt.Sprintf("%s=%s", k, v))
	}
}

func StartTime(ss float64) OptionFunc {
	return func(o *option) {
		o.ss = ss
	}
}

func Duration(t float64) OptionFunc {
	return func(o *option) {
		o.t = t
	}
}

func File(f string) OptionFunc {
	return func(o *option) {
		o.file = f
	}
}

func Format(f string) OptionFunc {
	return func(o *option) {
		o.f = f
	}
}

func VarStreamMap(s string) OptionFunc {
	return func(o *option) {
		o.var_stream_map = s
	}
}

// hls

func HlsSegmentType(value string) OptionFunc {
	return func(o *option) {
		o.hls_segment_type = value
	}
}

func HlsFlags(value string) OptionFunc {
	return func(o *option) {
		o.hls_flags = value
	}
}

func HlsPlaylistType(value string) OptionFunc {
	return func(o *option) {
		o.hls_playlist_type = value
	}
}
func HlsTime(value int32) OptionFunc {
	return func(o *option) {
		o.hls_time = value
	}
}
func MasterPlName(value string) OptionFunc {
	return func(o *option) {
		o.master_pl_name = value
	}
}
func HlsSegmentFilename(value string) OptionFunc {
	return func(o *option) {
		o.hls_segment_filename = value
	}
}

func HlsKeyInfoFile(f string) OptionFunc {
	return func(o *option) {
		o.hls_key_info_file = f
	}
}

// img

func Vframes(vframes int32) OptionFunc {
	return func(o *option) {
		o.vframes = vframes
	}
}
