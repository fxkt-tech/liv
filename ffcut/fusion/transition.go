package fusion

type Transition struct {
	Name string `json:"name,omitempty"`

	Duration int32  `json:"duration,omitempty"`
	Color    string `json:"color,omitempty"`
	// 淡入淡出同时处理音频
	WithAudio bool `json:"with_audio,omitempty"`
}
