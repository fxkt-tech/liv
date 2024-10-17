package conv

// 整型毫秒数转化为浮点数
func MillToF32(mill int32) float32 {
	return float32(mill) / 1000
}

// 浮点数转化为整型毫秒数
func F32ToMill(f float32) int32 {
	return int32(f * 1000)
}
