package fsugar

import "fmt"

type AssSubtitle struct {
	playResX  int32
	playResY  int32
	fontsize  int32
	marginv   int32
	fontname  string
	alignment int32
}

func NewAssSubtitle() *AssSubtitle {
	return &AssSubtitle{
		playResX:  1920,
		playResY:  1080,
		fontsize:  0,
		marginv:   10,
		fontname:  "",
		alignment: 6,
	}
}

func (s *AssSubtitle) SetPlayResX(playResX int32) *AssSubtitle {
	s.playResX = playResX
	return s
}

func (s *AssSubtitle) SetPlayResY(playResY int32) *AssSubtitle {
	s.playResY = playResY
	return s
}

func (s *AssSubtitle) SetFontSize(fontsize int32) *AssSubtitle {
	s.fontsize = fontsize
	return s
}

func (s *AssSubtitle) SetMarginV(marginv int32) *AssSubtitle {
	s.marginv = marginv
	return s
}

func (s *AssSubtitle) SetFontName(fontname string) *AssSubtitle {
	s.fontname = fontname
	return s
}

func (s *AssSubtitle) SetAlignment(alignment int32) *AssSubtitle {
	s.alignment = alignment
	return s
}

func (s *AssSubtitle) String() string {
	return fmt.Sprintf("PlayResX=%d,PlayResY=%d,Fontsize=%d,MarginV=%d,Fontname=%s,Alignment=%d",
		s.playResX, s.playResY, s.fontsize, s.marginv, s.fontname, s.alignment)
}
