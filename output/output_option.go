package output

type OptionFunc func(*option)

type option struct {
	maps                  []string // mean is -map.
	cv, ca                string   // cv is c:v, ca is c:a.
	metadatas             []string // mean is -metadata.
	threads               int32    // thread counts, default 4.
	max_muxing_queue_size int32    // queue size when muxing, default 4086.
	movflags              string   // location of mp4 moov.
	f                     string   // f is -f format.
	file                  string
}

func SetMap(m string) OptionFunc {
	return func(o *option) {
		o.maps = append(o.maps, m)
	}
}

func SetMetadata(md string) OptionFunc {
	return func(o *option) {
		o.metadatas = append(o.metadatas, md)
	}
}

func SetFile(f string) OptionFunc {
	return func(o *option) {
		o.file = f
	}
}
