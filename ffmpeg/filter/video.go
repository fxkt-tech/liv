package filter

import (
	"fmt"
	"strings"

	"github.com/fxkt-tech/liv/fftool/naming"
	"github.com/fxkt-tech/liv/pkg/math"
	"golang.org/x/exp/constraints"
)

type Expr interface {
	Numb | ~string
}

type Numb interface {
	~int32 | ~int
}

// 一个图像覆盖另一个图像
func Overlay[T Expr](dx, dy T, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("overlay=%v:%v", dx, dy), opts...),
	}
}

// 缩放
func Scale[T Expr](w, h T, opts ...FilterOpt) *SingleFilter {
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
		content: joinFilter(fmt.Sprintf("scale=%v:%v", ww, hh), opts...),
	}
}

// func UnsopportedSimple(fstr string) *SingleFilter {
// 	return &SingleFilter{
// 		name:    naming.Default.Gen(),
// 		content: fstr,
// 	}
// }

// 应用 3D LUT 调色文件
func Lut3D(file string) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("lut3d=file=%s:interp=tetrahedral", file),
	}
}

// ZScale 构造 FFmpeg zscale 视频滤镜，args 直接使用 FFmpeg 原生参数。
func ZScale(args string, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("zscale=%s", args), opts...),
	}
}

// Tonemap 构造 FFmpeg tonemap 视频滤镜，args 直接使用 FFmpeg 原生参数。
func Tonemap(args string, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("tonemap=%s", args), opts...),
	}
}

// Format 构造 FFmpeg format 视频滤镜。
func Format(pixFmt string, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("format=%s", pixFmt), opts...),
	}
}

// SetSAR 构造 FFmpeg setsar 视频滤镜。
func SetSAR[T Expr](sar T, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("setsar=%v", sar), opts...),
	}
}

// 绿幕抠像
func Chromakey(color string, similarity, blend float32, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name: naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf(
			"chromakey=%s:%.2f:%.2f",
			color, similarity, blend,
		), opts...),
	}
}

// 创建一个底版
func Color(c string, w, h int32, d float32, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name: naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf(
			"color=c=%s:s=%d*%d:d=%.2f:r=30",
			c, w, h, d,
		), opts...),
	}
}

// 裁切
func Crop[T Expr](w, h, x, y T, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name: naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf(
			"crop=%v:%v:%v:%v",
			w, h, x, y,
		), opts...),
	}
}

// 高斯模糊
func GBlur(sigma int32, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name: naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf(
			"gblur=sigma=%d", sigma,
		), opts...),
	}
}

// 视频淡入淡出
func Fade(t string, st, d float32, c string, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name: naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf(
			"fade=t=%s:st=%.2f:d=%.2f:c=%s",
			t, st, d, c,
		), opts...),
	}
}

// 字幕
func Subtitles(filename, fontsdir, forceStyle string, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name: naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf(
			"subtitles=f=%s:fontsdir=%s:force_style='%s'",
			filename, fontsdir, forceStyle,
		), opts...),
	}
}

// 视频帧显示时间戳
func SetPTS(expr string, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("setpts=%s", expr), opts...),
	}
}

// 截取某一时间段
func Trim(s, e float32, opts ...FilterOpt) *SingleFilter {
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
		content: joinFilter(fmt.Sprintf("trim%s%s", eqs, psstr), opts...),
	}
}

// 遮标
func Delogo[T Numb](x, y, w, h T, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name: naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("delogo=x=%v:y=%v:w=%v:h=%v",
			x+1, y+1, w-2, h-2,
		), opts...),
	}
}

// 绘制矩形区域
func DrawBox[T Expr](x, y, w, h T, color string, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name: naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("delogo=x=%v:y=%v:w=%v:h=%v:t=1:c=%s",
			x, y, w, h, color,
		), opts...),
	}
}

// 绘制填充矩形区域
func DrawBoxFill[T Expr](x, y, w, h T, color string, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("drawbox=%v:%v:%v:%v:%s:t=fill", x, y, w, h, color), opts...),
	}
}

// 色相与饱和度调整
func Hue(h float32, s float32, opts ...FilterOpt) *SingleFilter {
	content := fmt.Sprintf("hue=h=%g", h)
	if s != 0 {
		content += fmt.Sprintf(":s=%g", s)
	}
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: joinFilter(content, opts...),
	}
}

func Select(expr string, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("select=%s", expr), opts...),
	}
}

func FPS[N, D constraints.Integer | constraints.Float](fps *math.Rational[N, D], opts ...FilterOpt) *SingleFilter {
	var s string
	if fps.Den == 0 {
		s = "source_fps"
	} else {
		s = fmt.Sprintf("%v/%v", fps.Num, fps.Den)
	}
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("fps=fps=%s", s), opts...),
	}
}

func Tile(xlen, ylen int32, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("tile=%d*%d", xlen, ylen), opts...),
	}
}

// multi

// 视频流复制成多份
func Split(n int, opts ...FilterOpt) *MultiFilter {
	return &MultiFilter{
		name:    naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("split=%d", n), opts...),
		counts:  n,
	}
}
