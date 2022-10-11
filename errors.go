package liv

import "errors"

var (
	ErrParamsInvalid  = errors.New("params is invalid")
	ErrParamsInvalid2 = errors.New("interval required when frame_type is 1(normal frame)")
)
