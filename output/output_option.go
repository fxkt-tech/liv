package output

import "fmt"

type OptionFunc func(*option)

type option struct {
	maps                  []string // mean is -map.
	cv, ca                string   // cv is c:v, ca is c:a.
	bv, ba                int32    // bv is b:v, ba is b:a.
	metadatas             []string // mean is -metadata.
	threads               int32    // thread counts, default 4.
	max_muxing_queue_size int32    // queue size when muxing, default 4086.
	movflags              string   // location of mp4 moov.
	ss                    float64  // ss is start_time.
	t                     float64  // t is duration.
	f                     string   // f is -f format.
	file                  string

	// hls configs
	hls_segment_type     string
	hls_flags            string
	hls_playlist_type    string
	hls_time             int32
	master_pl_name       string
	hls_segment_filename string

	// img
	vframes int32
}

func SetMap(m string) OptionFunc {
	return func(o *option) {
		o.maps = append(o.maps, m)
	}
}

func SetMovFlags(flag string) OptionFunc {
	return func(o *option) {
		o.movflags = flag
	}
}

func SetThread(n int32) OptionFunc {
	return func(o *option) {
		if n > 0 {
			o.threads = n
		}
	}
}

func SetMaxMuxingQueueSize(sz int32) OptionFunc {
	return func(o *option) {
		o.max_muxing_queue_size = sz
	}
}

func SetVideoCoder(cv string) OptionFunc {
	return func(o *option) {
		o.cv = cv
	}
}

func SetAudioCoder(ca string) OptionFunc {
	return func(o *option) {
		o.ca = ca
	}
}

func SetVideoBitrate(bv int32) OptionFunc {
	return func(o *option) {
		o.bv = bv
	}
}

func SetAudioBitrate(ba int32) OptionFunc {
	return func(o *option) {
		o.ba = ba
	}
}

func SetMetadata(k, v string) OptionFunc {
	return func(o *option) {
		o.metadatas = append(o.metadatas, fmt.Sprintf("%s=%s", k, v))
	}
}

func SetStartTime(ss float64) OptionFunc {
	return func(o *option) {
		o.ss = ss
	}
}

func SetDuration(t float64) OptionFunc {
	return func(o *option) {
		o.t = t
	}
}

func SetFile(f string) OptionFunc {
	return func(o *option) {
		o.file = f
	}
}

func SetFormat(f string) OptionFunc {
	return func(o *option) {
		o.f = f
	}
}

// hls

func SetHlsSegmentType(value string) OptionFunc {
	return func(o *option) {
		o.hls_segment_type = value
	}
}

func SetHlsFlags(value string) OptionFunc {
	return func(o *option) {
		o.hls_flags = value
	}
}

func SetHlsPlaylistType(value string) OptionFunc {
	return func(o *option) {
		o.hls_playlist_type = value
	}
}
func SetHlsTime(value int32) OptionFunc {
	return func(o *option) {
		o.hls_time = value
	}
}
func SetMasterPlName(value string) OptionFunc {
	return func(o *option) {
		o.master_pl_name = value
	}
}
func SetHlsSegmentFilename(value string) OptionFunc {
	return func(o *option) {
		o.hls_segment_filename = value
	}
}

// img

func SetVframes(vframes int32) OptionFunc {
	return func(o *option) {
		o.vframes = vframes
	}
}
