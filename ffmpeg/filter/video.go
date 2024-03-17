package filter

import (
	"fmt"
	"strings"

	"github.com/fxkt-tech/liv/internal/math"
	"github.com/fxkt-tech/liv/internal/naming"
	"golang.org/x/exp/constraints"
)

type LogoPos string

const (
	LogoTopLeft     LogoPos = "TopLeft"
	LogoTopRight    LogoPos = "TopRight"
	LogoBottomRight LogoPos = "BottomRight"
	LogoBottomLeft  LogoPos = "BottomLeft"
)

// 贴logo
func Logo(dx, dy int64, pos LogoPos) Filter {
	var content string
	switch pos {
	case LogoTopLeft:
		content = fmt.Sprintf("overlay=%d:%d", dx, dy)
	case LogoTopRight:
		content = fmt.Sprintf("overlay=W-w-%d:%d", dx, dy)
	case LogoBottomRight:
		content = fmt.Sprintf("overlay=W-w-%d:H-h-%d", dx, dy)
	case LogoBottomLeft:
		content = fmt.Sprintf("overlay=%d:H-h-%d", dx, dy)
	}
	return &single{
		name:    naming.Default.Gen(),
		content: content,
	}
}

// 一个图像覆盖另一个图像
func Overlay[T constraints.Signed | string](dx, dy T) Filter {
	return &single{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("overlay=%v:%v", dx, dy),
	}
}

// 一个图像覆盖另一个图像（可激活某一时间段）
func OverlayWithEnable[T constraints.Signed | string](dx, dy T, enable string) Filter {
	return &single{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("overlay=%v:%v:enable='%s'", dx, dy, enable),
	}
}

// 缩放
func Scale[T int32 | int | string](w, h T) Filter {
	var ww, hh any = w, h
	switch ww.(type) {
	case int32:
		ww, hh = math.CeilEven(ww.(int32)), math.CeilEven(hh.(int32))
	case int:
		ww, hh = math.CeilEven(ww.(int)), math.CeilEven(hh.(int))
	case string:
		ww, hh = w, h
	}
	return &single{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("scale=%v:%v", ww, hh),
	}
}

// 绿幕抠像
func Chromakey(color string, similarity, blend float32) Filter {
	return &single{
		name: naming.Default.Gen(),
		content: fmt.Sprintf(
			"chromakey=%s:%.2f:%.2f",
			color, similarity, blend,
		),
	}
}

// 创建一个底版
func Color(c string, w, h int32, d float32) Filter {
	return &single{
		name: naming.Default.Gen(),
		content: fmt.Sprintf(
			"color=c=%s:s=%d*%d:d=%.2f",
			c, w, h, d,
		),
	}
}

// 裁切
func Crop(x, y, w, h int32) Filter {
	return &single{
		name: naming.Default.Gen(),
		content: fmt.Sprintf(
			"crop=%d:%d:%d:%d",
			x, y, w, h,
		),
	}
}

// 视频帧显示时间戳
func SetPTS(expr string) Filter {
	return &single{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("setpts=%s", expr),
	}
}

// 视频流复制成多份
func Split(n int) Filter {
	return &multiple{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("split=%d", n),
		counts:  n,
	}
}

// 截取某一时间段
func Trim(s, e float64) Filter {
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
	return &single{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("trim%s%s", eqs, psstr),
	}
}

// 擦除logo
func Delogo(x, y, w, h int64) Filter {
	return &single{
		name: naming.Default.Gen(),
		content: fmt.Sprintf("delogo=%d:%d:%d:%d",
			x+1, y+1, w-2, h-2,
		),
	}
}

func Select(expr string) Filter {
	return &single{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("select=%s", expr),
	}
}

func FPS[N, D constraints.Integer | constraints.Float](fps *math.Rational[N, D]) Filter {
	var s string
	if fps.Den == 0 {
		s = "source_fps"
	} else {
		s = fmt.Sprintf("%v/%v", fps.Num, fps.Den)
	}
	return &single{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("fps=fps=%s", s),
	}
}

func Tile(xlen, ylen int32) Filter {
	return &single{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("tile=%d*%d", xlen, ylen),
	}
}
