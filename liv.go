package liv

import "context"

type Transcoder interface {
	Transcode(context.Context, *TranscodeParams) error
}

type VideoFilter interface {
}
