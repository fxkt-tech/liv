package codec

const (
	Copy = "copy"
	Nop  = "nop"

	// audio
	AAC    = "aac"
	FDKAAC = "libfdk_aac"

	// video
	X264    = "libx264"
	WZ264   = "libwz264"
	X265    = "libx265"
	WZ265   = "libwz265"
	MP3Lame = "libmp3lame"
	MJPEG   = "mjpeg"

	// container
	HLS  = "hls"
	Dash = "dash"
	MP4  = "mp4"
	MP3  = "mp3"
	JPEG = "jpeg"
	JPG  = "jpg"
	PNG  = "png"
	WEBP = "webp"
)
