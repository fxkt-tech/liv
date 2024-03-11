package filter

import (
	"fmt"
	"strings"

	"github.com/fxkt-tech/liv/internal/math"
	"golang.org/x/exp/constraints"
)

type Filter interface {
	Name(int) string
	Index() int
	Copy(int) Filter
	String() string
	Use(...Filter) Filter
}

type CommonFilter struct {
	name    string
	index   int
	counts  int
	content string
	uses    []Filter
}

func (cf *CommonFilter) Name(index int) string {
	if cf.name == "" {
		return ""
	}
	if cf.counts == 0 {
		return fmt.Sprintf("[%s]", cf.name)
	}
	return fmt.Sprintf("[%s%d]", cf.name, index)
}

func (cf *CommonFilter) Index() int {
	return cf.index
}

func (cf *CommonFilter) Copy(index int) Filter {
	return &CommonFilter{
		name:    cf.name,
		index:   index,
		counts:  cf.counts,
		content: cf.content,
		uses:    cf.uses,
	}
}

func (cf *CommonFilter) String() string {
	fls := make([]string, len(cf.uses))
	for i, fl := range cf.uses {
		if fl != nil {
			fls[i] = fl.Name(fl.Index())
		}
	}
	var names []string
	for i := 0; i < cf.counts; i++ {
		names = append(names, cf.Name(i))
	}
	return fmt.Sprintf("%s%s%s", strings.Join(fls, ""), cf.content, strings.Join(names, ""))
}

func (cf *CommonFilter) Use(fls ...Filter) Filter {
	cf.uses = append(cf.uses, fls...)
	return cf
}

// type of filter

type Stream string

const (
	StreamAudio Stream = "a"
	StreamVideo Stream = "v"
)

// SelectStream used for input stream selection.
func SelectStream(idx int, s Stream, must bool) Filter {
	var qm string
	if !must {
		qm = "?"
	}
	return &CommonFilter{
		name: fmt.Sprintf("%d:%s%s", idx, s, qm),
	}
}

type LogoPos string

const (
	LogoTopLeft     LogoPos = "TopLeft"
	LogoTopRight    LogoPos = "TopRight"
	LogoBottomRight LogoPos = "BottomRight"
	LogoBottomLeft  LogoPos = "BottomLeft"
)

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
	return &CommonFilter{
		name:    name,
		content: content,
		counts:  1,
	}
}

func Overlay[T int32 | int | string](name string, dx, dy T) Filter {
	return &CommonFilter{
		name:    name,
		content: fmt.Sprintf("overlay=%v:%v", dx, dy),
		counts:  1,
	}
}

func OverlayWithEnable[T int32 | int | string](name string, dx, dy T, enable string) Filter {
	return &CommonFilter{
		name:    name,
		content: fmt.Sprintf("overlay=%v:%v:enable='%s'", dx, dy, enable),
		counts:  1,
	}
}

func Scale[T int32 | int | string](name string, w, h T) Filter {
	var ww, hh any = w, h
	switch ww.(type) {
	case int32, int:
		ww, hh = math.CeilOddInt32(ww.(int32)), math.CeilOddInt32(hh.(int32))
	case string:
		ww, hh = w, h
	}
	return &CommonFilter{
		name:    name,
		content: fmt.Sprintf("scale=%v:%v", ww, hh),
		counts:  1,
	}
}

func Chromakey(name string, color string, similarity, blend float32) Filter {
	return &CommonFilter{
		name: name,
		content: fmt.Sprintf(
			"chromakey=%s:%.2f:%.2f",
			color, similarity, blend,
		),
		counts: 1,
	}
}

func Color(name string, c string, w, h int32, d float32) Filter {
	return &CommonFilter{
		name: name,
		content: fmt.Sprintf(
			"color=c=%s:s=%d*%d:d=%.2f",
			c, w, h, d,
		),
		counts: 1,
	}
}

func Crop(name string, x, y, w, h int32) Filter {
	return &CommonFilter{
		name: name,
		content: fmt.Sprintf(
			"crop=%d:%d:%d:%d",
			x, y, w, h,
		),
		counts: 1,
	}
}

func SetPTS(name string, expr string) Filter {
	return &CommonFilter{
		name:    name,
		content: fmt.Sprintf("setpts=%s", expr),
		counts:  1,
	}
}

func ASetPTS(name string, expr string) Filter {
	return &CommonFilter{
		name:    name,
		content: fmt.Sprintf("asetpts=%s", expr),
		counts:  1,
	}
}

func Split(name string, n int) Filter {
	return &CommonFilter{
		name:    name,
		content: fmt.Sprintf("split=%d", n),
		counts:  n,
	}
}

func ASplit(name string, n int) Filter {
	return &CommonFilter{
		name:    name,
		content: fmt.Sprintf("asplit=%d", n),
		counts:  n,
	}
}

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
	return &CommonFilter{
		name:    name,
		content: fmt.Sprintf("trim%s%s", eqs, psstr),
		counts:  1,
	}
}

func Delogo(name string, x, y, w, h int64) Filter {
	return &CommonFilter{
		name: name,
		content: fmt.Sprintf("delogo=%d:%d:%d:%d",
			x+1, y+1, w-2, h-2,
		),
		counts: 1,
	}
}

func Select(name, expr string) Filter {
	return &CommonFilter{
		name:    name,
		content: fmt.Sprintf("select=%s", expr),
		counts:  1,
	}
}

func FPS[N, D constraints.Integer | constraints.Float](name string, fps *math.Rational[N, D]) Filter {
	var s string
	if fps.Den == 0 {
		s = "source_fps"
	} else {
		s = fmt.Sprintf("%v/%v", fps.Num, fps.Den)
	}
	return &CommonFilter{
		name:    name,
		content: fmt.Sprintf("fps=fps=%s", s),
		counts:  1,
	}
}

func Tile(name string, xlen, ylen int32) Filter {
	return &CommonFilter{
		name:    name,
		content: fmt.Sprintf("tile=%d*%d", xlen, ylen),
		counts:  1,
	}
}

func AMix(name string, inputs int32) Filter {
	return &CommonFilter{
		name:    name,
		content: fmt.Sprintf("amix=inputs=%d", inputs),
		counts:  1,
	}
}

func Loudnorm(name string, i, tp int32) Filter {
	return &CommonFilter{
		name:    name,
		content: fmt.Sprintf("loudnorm=I=%d:TP=%d", i, tp),
		counts:  1,
	}
}

func ADelay(name string, delays int32) Filter {
	return &CommonFilter{
		name:    name,
		content: fmt.Sprintf("adelay=delays=%ds:all=1", delays),
		counts:  1,
	}
}

// filter slice

type Filters []Filter

func (filters Filters) Params() (params []string) {
	txt := filters.String()
	if txt == "" {
		return
	}
	params = append(params, "-filter_complex", txt)
	return
}

func (filters Filters) String() string {
	var params []string
	for _, filter := range filters {
		params = append(params, filter.String())
	}
	return strings.Join(params, ";")
}
