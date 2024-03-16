package filter

import (
	"fmt"
	"strings"

	"github.com/fxkt-tech/liv/internal/math"
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
func Logo(name string, dx, dy int64, pos LogoPos) Filter {
	var content string
	switch pos {
	case LogoTopLeft:
		content = fmt.Sprintf("overlay=%d:y=%d", dx, dy)
	case LogoTopRight:
		content = fmt.Sprintf("overlay=W-w-%d:y=%d", dx, dy)
	case LogoBottomRight:
		content = fmt.Sprintf("overlay=W-w-%d:y=H-h-%d", dx, dy)
	case LogoBottomLeft:
		content = fmt.Sprintf("overlay=%d:y=H-h-%d", dx, dy)
	}
	return &single{
		name:    name,
		content: content,
	}
}

// 一个图像覆盖另一个图像
func Overlay[T int32 | int | string](name string, dx, dy T) Filter {
	return &single{
		name:    name,
		content: fmt.Sprintf("overlay=%v:%v", dx, dy),
	}
}

// 一个图像覆盖另一个图像（可激活某一时间段）
func OverlayWithEnable[T int32 | int | string](name string, dx, dy T, enable string) Filter {
	return &single{
		name:    name,
		content: fmt.Sprintf("overlay=%v:%v:enable='%s'", dx, dy, enable),
	}
}

// 缩放
func Scale[T int32 | int | string](name string, w, h T) Filter {
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
		name:    name,
		content: fmt.Sprintf("scale=%v:%v", ww, hh),
	}
}

// 绿幕抠像
func Chromakey(name string, color string, similarity, blend float32) Filter {
	return &single{
		name: name,
		content: fmt.Sprintf(
			"chromakey=%s:%.2f:%.2f",
			color, similarity, blend,
		),
	}
}

// 创建一个底版
func Color(name string, c string, w, h int32, d float32) Filter {
	return &single{
		name: name,
		content: fmt.Sprintf(
			"color=c=%s:s=%d*%d:d=%.2f",
			c, w, h, d,
		),
	}
}

// 裁切
func Crop(name string, x, y, w, h int32) Filter {
	return &single{
		name: name,
		content: fmt.Sprintf(
			"crop=%d:%d:%d:%d",
			x, y, w, h,
		),
	}
}

// 视频帧显示时间戳
func SetPTS(name string, expr string) Filter {
	return &single{
		name:    name,
		content: fmt.Sprintf("setpts=%s", expr),
	}
}

// 音频帧显示时间戳
func ASetPTS(name string, expr string) Filter {
	return &single{
		name:    name,
		content: fmt.Sprintf("asetpts=%s", expr),
	}
}

// 视频流复制成多份
func Split(name string, n int) Filter {
	return &multiple{
		name:    name,
		content: fmt.Sprintf("split=%d", n),
		counts:  n,
	}
}

// 截取某一时间段
func Trim(name string, s, e float64) Filter {
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
		name:    name,
		content: fmt.Sprintf("trim%s%s", eqs, psstr),
	}
}

// 擦除logo
func Delogo(name string, x, y, w, h int64) Filter {
	return &single{
		name: name,
		content: fmt.Sprintf("delogo=%d:%d:%d:%d",
			x+1, y+1, w-2, h-2,
		),
	}
}

func Select(name, expr string) Filter {
	return &single{
		name:    name,
		content: fmt.Sprintf("select=%s", expr),
	}
}

func FPS[N, D constraints.Integer | constraints.Float](name string, fps *math.Rational[N, D]) Filter {
	var s string
	if fps.Den == 0 {
		s = "source_fps"
	} else {
		s = fmt.Sprintf("%v/%v", fps.Num, fps.Den)
	}
	return &single{
		name:    name,
		content: fmt.Sprintf("fps=fps=%s", s),
	}
}

func Tile(name string, xlen, ylen int32) Filter {
	return &single{
		name:    name,
		content: fmt.Sprintf("tile=%d*%d", xlen, ylen),
	}
}
