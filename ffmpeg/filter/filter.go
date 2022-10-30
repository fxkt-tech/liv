package filter

import (
	"fmt"
	"strings"

	"github.com/fxkt-tech/liv/internal/math"
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
		fls[i] = fl.Name(fl.Index())
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
func SelectStream(idx int, s Stream, must bool) string {
	var qm string
	if !must {
		qm = "?"
	}
	return fmt.Sprintf("%d:%s%s", idx, s, qm)
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

func Scale(name string, w, h int32) Filter {
	return &CommonFilter{
		name: name,
		content: fmt.Sprintf("scale=%d:%d",
			math.CeilOddInt32(w),
			math.CeilOddInt32(h),
		),
		counts: 1,
	}
}

func Split(name string, n int) Filter {
	return &CommonFilter{
		name:    name,
		content: fmt.Sprintf("split=%d", n),
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

func FPS(name string, fps *math.Rational[int32]) Filter {
	return &CommonFilter{
		name:    name,
		content: fmt.Sprintf("fps=fps=%d/%d", fps.Num, fps.Den),
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
