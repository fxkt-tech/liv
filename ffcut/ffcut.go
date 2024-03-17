package ffcut

import "github.com/fxkt-tech/liv/internal/sugar"

type FFcut struct {
	//
}

func New(opts ...Option) *FFcut {
	fc := &FFcut{}
	sugar.Range(opts, func(o Option) { o(fc) })
	return fc
}
