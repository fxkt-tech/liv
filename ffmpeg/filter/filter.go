package filter

import (
	"fmt"
	"strings"

	"github.com/fxkt-tech/liv/ffmpeg/stream"
)

type Filter interface {
	stream.Streamer
	String() string
}

// 单输出滤镜
type SingleFilter struct {
	name    string
	content string
	uses    []stream.Streamer
}

func (s *SingleFilter) Name(_ stream.PosFrom) string {
	if s.name == "" {
		return ""
	}
	return fmt.Sprintf("[%s]", s.name)
}

func (s *SingleFilter) S() stream.Streamer { return s }

func (s *SingleFilter) String() string {
	fls := make([]string, len(s.uses))
	for i, fl := range s.uses {
		if fl != nil {
			fls[i] = fl.Name(stream.PosFromFilter)
		}
	}
	return fmt.Sprintf("%s%s%s", strings.Join(fls, ""), s.content, s.Name(stream.PosFromFilter))
}

func (s *SingleFilter) Use(streams ...stream.Streamer) *SingleFilter {
	s.uses = append(s.uses, streams...)
	return s
}

// 多输出滤镜
type MultiFilter struct {
	name    string
	counts  int
	content string
	uses    []stream.Streamer
}

func (s *MultiFilter) Name(_ stream.PosFrom) string {
	if s.name == "" {
		return ""
	}
	return fmt.Sprintf("[%s]", s.name)
}

// 选择一个
func (s *MultiFilter) S(i int) stream.Streamer {
	name := fmt.Sprintf("[%s_%d]", s.name, i)
	return stream.StreamImpl(name)
}

func (s *MultiFilter) String() string {
	fls := make([]string, len(s.uses))
	for i, fl := range s.uses {
		if fl != nil {
			fls[i] = fl.Name(stream.PosFromFilter)
		}
	}
	var names []string
	for i := 0; i < s.counts; i++ {
		names = append(names, fmt.Sprintf("[%s_%d]", s.name, i))
	}
	return fmt.Sprintf("%s%s%s", strings.Join(fls, ""), s.content, strings.Join(names, ""))
}

func (s *MultiFilter) Use(streams ...stream.Streamer) *MultiFilter {
	s.uses = append(s.uses, streams...)
	return s
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
