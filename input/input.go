package input

import (
	"strconv"
	"strings"
)

// Input is 输入.
type Input struct {
	i   string   // 输入的媒体文件
	ss  int64    // 媒体文件选择的起始时间点
	t   int64    //从起始时间开始的持续时间
	ext []string // 额外字段
}

func (i *Input) Params() (params []string) {
	if i.ss != 0 {
		params = append(params, "-ss", strconv.FormatInt(i.ss, 10))
	}
	if i.t != 0 {
		params = append(params, "-t", strconv.FormatInt(i.t, 10))
	}
	params = append(params, "-i", i.i)
	return
}

func (i *Input) String() string {
	return strings.Join(i.Params(), " ")
}
