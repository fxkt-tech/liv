package ffmpeg

type LogLevel string

func (l LogLevel) String() string { return string(l) }

const (
	LogLevelQuiet LogLevel = "quiet"
	LogLevelError LogLevel = "error"
)
