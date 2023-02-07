package util

func FixPixelLen(l int32) int32 {
	if l == 0 {
		return -2
	}
	return l
}
