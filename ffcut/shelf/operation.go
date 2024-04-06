package shelf

// 元素操作
type OpParams interface{}

// 操作
type Operation struct {
	Type   string   `json:"type,omitempty"`
	Params OpParams `json:"params,omitempty"`
}

// 旋转
type OpParamsImageRotate struct {
	Angle int32 `json:"angle"`
}

func WithOpParamsImageRotate(angle int32) *Operation {
	return &Operation{
		Type: "image_rotate",
		Params: &OpParamsImageRotate{
			Angle: angle,
		},
	}
}

// 调整音量
type OpParamsAudioVolumes struct {
	All int32 `json:"all"`
}

func WithOpParamsAudioVolumes(all int32) *Operation {
	return &Operation{
		Type: "audio_volumes",
		Params: &OpParamsAudioVolumes{
			All: all,
		},
	}
}
