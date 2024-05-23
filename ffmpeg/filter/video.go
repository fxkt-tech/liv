package filter

import (
	"fmt"
	"strings"

	"github.com/fxkt-tech/liv/internal/math"
	"github.com/fxkt-tech/liv/internal/naming"
	"golang.org/x/exp/constraints"
)

type Expr interface {
	Numb | ~string
}

type Numb interface {
	~int32 | ~int
}

// 一个图像覆盖另一个图像
func Overlay[T Expr](dx, dy T) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("overlay=%v:%v", dx, dy),
	}
}

// 一个图像覆盖另一个图像（可激活某一时间段）
func OverlayWithEnable[T Expr](dx, dy T, enable string) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("overlay=%v:%v:enable='%s'", dx, dy, enable),
	}
}

func OverlayWithTime[T Expr](dx, dy T, t float32) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("overlay=%v:%v:t=%f", dx, dy, t),
	}
}

// 缩放
func Scale[T Expr](w, h T) *SingleFilter {
	var ww, hh any = w, h
	switch ww.(type) {
	case int32:
		ww, hh = math.CeilEven(ww.(int32)), math.CeilEven(hh.(int32))
	case int:
		ww, hh = math.CeilEven(ww.(int)), math.CeilEven(hh.(int))
	case string:
		ww, hh = w, h
	}
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("scale=%v:%v", ww, hh),
	}
}

// func UnsopportedSimple(fstr string) *SingleFilter {
// 	return &SingleFilter{
// 		name:    naming.Default.Gen(),
// 		content: fstr,
// 	}
// }

// 绿幕抠像
func Chromakey(color string, similarity, blend float32) *SingleFilter {
	return &SingleFilter{
		name: naming.Default.Gen(),
		content: fmt.Sprintf(
			"chromakey=%s:%.2f:%.2f",
			color, similarity, blend,
		),
	}
}

// 创建一个底版
func Color(c string, w, h int32, d float32) *SingleFilter {
	return &SingleFilter{
		name: naming.Default.Gen(),
		content: fmt.Sprintf(
			"color=c=%s:s=%d*%d:d=%.2f:r=30",
			c, w, h, d,
		),
	}
}

// 裁切
func Crop[T Expr](x, y, w, h T) *SingleFilter {
	return &SingleFilter{
		name: naming.Default.Gen(),
		content: fmt.Sprintf(
			"crop=%v:%v:%v:%v",
			x, y, w, h,
		),
	}
}

// 视频帧显示时间戳
func SetPTS(expr string) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("setpts=%s", expr),
	}
}

// 截取某一时间段
func Trim(s, e float32) *SingleFilter {
	var ps []string
	if s != 0 {
		ps = append(ps, fmt.Sprintf("start=%f", s))
	}
	if e != 0 {
		ps = append(ps, fmt.Sprintf("end=%f", e))
	}
	var eqs string
	var psstr string = strings.Join(ps, ":")
	if psstr != "" {
		eqs = "="
	}
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("trim%s%s", eqs, psstr),
	}
}

// Deprecated: 擦除logo
func Delogo_Old(x, y, w, h int32) *SingleFilter {
	return &SingleFilter{
		name: naming.Default.Gen(),
		content: fmt.Sprintf("delogo=%d:%d:%d:%d",
			x+1, y+1, w-2, h-2,
		),
	}
}

// 遮标
func Delogo[T Numb](x, y, w, h T) *SingleFilter {
	return &SingleFilter{
		name: naming.Default.Gen(),
		content: fmt.Sprintf("delogo=x=%v:y=%v:w=%v:h=%v",
			x+1, y+1, w-2, h-2,
		),
	}
}

func Select(expr string) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("select=%s", expr),
	}
}

func FPS[N, D constraints.Integer | constraints.Float](fps *math.Rational[N, D]) *SingleFilter {
	var s string
	if fps.Den == 0 {
		s = "source_fps"
	} else {
		s = fmt.Sprintf("%v/%v", fps.Num, fps.Den)
	}
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("fps=fps=%s", s),
	}
}

func Tile(xlen, ylen int32) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("tile=%d*%d", xlen, ylen),
	}
}

// multi

// 视频流复制成多份
func Split(n int) *MultiFilter {
	return &MultiFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("split=%d", n),
		counts:  n,
	}
}
