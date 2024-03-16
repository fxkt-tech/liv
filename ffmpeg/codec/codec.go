package codec

const (
	Copy = "copy"
	Nope = "nope"

	// audio
	AAC     = "aac"
	FDKAAC  = "libfdk_aac"
	MP3Lame = "libmp3lame"

	// video
	X264  = "libx264"
	WZ264 = "libwz264"
	X265  = "libx265"
	WZ265 = "libwz265"
	VP9   = "libvpx-vp9"

	// image
	MJPEG = "mjpeg"

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
