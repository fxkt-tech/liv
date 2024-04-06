package shelf

type ShelfOption func(*TrackData)

func WithStageSize(w, h int32) ShelfOption {
	return func(td *TrackData) {
		td.stageWidth, td.stageHeight = w, h
	}
}
