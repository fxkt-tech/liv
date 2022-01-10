package filter

import (
	"fmt"
	"strings"

	"fxkt.tech/ffmpeg/internal/math"
)

type Stream string

const (
	StreamAudio Stream = "a"
	StreamVideo Stream = "v"
)

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

func Logo(dx, dy int64, pos LogoPos) string {
	switch pos {
	case LogoTopLeft:
		return fmt.Sprintf("overlay=%d:y=%d", dx, dy)
	case LogoTopRight:
		return fmt.Sprintf("overlay=W-w-%d:y=%d", dx, dy)
	case LogoBottomRight:
		return fmt.Sprintf("overlay=W-w-%d:y=H-h-%d", dx, dy)
	case LogoBottomLeft:
		return fmt.Sprintf("overlay=%d:y=H-h-%d", dx, dy)
	}
	return ""
}

func Scale(w, h int32) string {
	return fmt.Sprintf("scale=%d:%d",
		math.CeilOddInt32(w),
		math.CeilOddInt32(h),
	)
}

func Split(n int) string {
	return fmt.Sprintf("split=%d", n)
}

func Trim(s, e float64) string {
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
	return fmt.Sprintf("trim%s%s", eqs, psstr)
}

func Delogo(x, y, w, h int64) string {
	return fmt.Sprintf("delogo=%d:%d:%d:%d",
		x+1, y+1, w-2, h-2,
	)
}
