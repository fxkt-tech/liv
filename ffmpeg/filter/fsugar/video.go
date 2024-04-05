package fsugar

import (
	"fmt"

	"github.com/fxkt-tech/liv/ffmpeg/filter"
)

const (
	LogoPosTopLeft     = "TopLeft"
	LogoPosTopRight    = "TopRight"
	LogoPosBottomRight = "BottomRight"
	LogoPosBottomLeft  = "BottomLeft"
)

// logo位置
func LogoPos[T filter.Expr](dx, dy T, pos string) (string, string) {
	switch pos {
	case LogoPosTopLeft:
		return fmt.Sprintf("%v", dx), fmt.Sprintf("%v", dy)
	case LogoPosTopRight:
		return fmt.Sprintf("W-w-%v", dx), fmt.Sprintf("%v", dy)
	case LogoPosBottomRight:
		return fmt.Sprintf("W-w-%v", dx), fmt.Sprintf("H-h-%v", dy)
	case LogoPosBottomLeft:
		return fmt.Sprintf("%v", dx), fmt.Sprintf("H-h-%v", dy)
	}
	return "", ""
}
