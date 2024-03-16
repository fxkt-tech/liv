package filter

import (
	"fmt"
	"strings"

	"github.com/fxkt-tech/liv/ffmpeg/stream"
)

type Filter interface {
	stream.Streamer
	Get(int) stream.Streamer
	Copy(int) Filter
	String() string
	Use(...stream.Streamer) Filter
}

// 单输出滤镜
type single struct {
	name    string
	content string
	uses    []stream.Streamer
}

func (s *single) Name() string {
	if s.name == "" {
		return ""
	}
	return fmt.Sprintf("[%s]", s.name)
}

func (s *single) Get(i int) stream.Streamer { return s }

func (s *single) Copy(index int) Filter {
	return &single{
		name:    s.name,
		content: s.content,
		uses:    s.uses,
	}
}

func (s *single) String() string {
	fls := make([]string, len(s.uses))
	for i, fl := range s.uses {
		if fl != nil {
			fls[i] = fl.Name()
		}
	}
	return fmt.Sprintf("%s%s%s", strings.Join(fls, ""), s.content, s.Name())
}

func (s *single) Use(streams ...stream.Streamer) Filter {
	s.uses = append(s.uses, streams...)
	return s
}

// 多输出滤镜
type multiple struct {
	name    string
	counts  int
	content string
	uses    []stream.Streamer
}

func (s *multiple) Name() string {
	if s.name == "" {
		return ""
	}
	return fmt.Sprintf("[%s]", s.name)
}

func (s *multiple) Get(i int) stream.Streamer {
	name := fmt.Sprintf("[%s_%d]", s.name, i)
	return stream.StreamImpl(name)
}

func (s *multiple) Copy(index int) Filter {
	return &multiple{
		name:    s.name,
		counts:  s.counts,
		content: s.content,
		uses:    s.uses,
	}
}

func (s *multiple) String() string {
	fls := make([]string, len(s.uses))
	for i, fl := range s.uses {
		if fl != nil {
			fls[i] = fl.Name()
		}
	}
	var names []string
	for i := 0; i < s.counts; i++ {
		names = append(names, fmt.Sprintf("[%s_%d]", s.name, i))
	}
	return fmt.Sprintf("%s%s%s", strings.Join(fls, ""), s.content, strings.Join(names, ""))
}

func (s *multiple) Use(streams ...stream.Streamer) Filter {
	s.uses = append(s.uses, streams...)
	return s
}

// filter slice

type Filters []Filter

// func (filters Filters) RefInput() Filters {

// 	return  filters
// }

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
