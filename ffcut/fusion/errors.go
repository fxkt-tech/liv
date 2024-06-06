package fusion

import "errors"

var (
	// 轨道类型不存在
	ErrTrackTypeNotFound = errors.New("track type not found")
	// 元素与轨道不匹配
	ErrTrackItemTypeNotMatch = errors.New("track and track_type not match")
)
