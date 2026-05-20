package filter

import (
	"fmt"
	"strings"
)

type FilterOpt string

func WithKV(k string, v any) FilterOpt {
	if k == "" || v == nil {
		return ""
	}
	sv := fmt.Sprint(v)
	if sv == "" {
		return ""
	}
	return FilterOpt(fmt.Sprintf("%s=%s", k, sv))
}

func WithKVs(kvs map[string]any) []FilterOpt {
	opts := make([]FilterOpt, 0, len(kvs))
	for k, v := range kvs {
		if opt := WithKV(k, v); opt != "" {
			opts = append(opts, opt)
		}
	}
	return opts
}

func WithEnable(expr string) FilterOpt {
	if expr == "" {
		return ""
	}
	return FilterOpt(fmt.Sprintf("enable='%s'", expr))
}

func WithTimeMode(v string) FilterOpt    { return WithKV("t", v) }
func WithShortest(v int) FilterOpt       { return WithKV("shortest", v) }
func WithFormat(v string) FilterOpt      { return WithKV("format", v) }
func WithAlpha(v string) FilterOpt       { return WithKV("alpha", v) }
func WithFOAR(v string) FilterOpt        { return WithKV("force_original_aspect_ratio", v) }
func WithFlags(v string) FilterOpt       { return WithKV("flags", v) }
func WithEval(v string) FilterOpt        { return WithKV("eval", v) }
func WithW[T Expr](v T) FilterOpt        { return WithKV("w", v) }
func WithH[T Expr](v T) FilterOpt        { return WithKV("h", v) }
func WithX[T Expr](v T) FilterOpt        { return WithKV("x", v) }
func WithY[T Expr](v T) FilterOpt        { return WithKV("y", v) }
func WithColor(v string) FilterOpt       { return WithKV("color", v) }
func WithSigma(v int32) FilterOpt        { return WithKV("sigma", v) }
func WithSteps(v int32) FilterOpt        { return WithKV("steps", v) }
func WithSimilarity(v float32) FilterOpt { return WithKV("similarity", v) }
func WithBlend(v float32) FilterOpt      { return WithKV("blend", v) }
func WithStart(v float32) FilterOpt      { return WithKV("start", v) }
func WithEnd(v float32) FilterOpt        { return WithKV("end", v) }
func WithInputs(v int32) FilterOpt       { return WithKV("inputs", v) }
func WithDuration(v string) FilterOpt    { return WithKV("duration", v) }
func WithNormalize(v int32) FilterOpt    { return WithKV("normalize", v) }
func WithDelays(v string) FilterOpt      { return WithKV("delays", v) }
func WithAll(v int32) FilterOpt          { return WithKV("all", v) }
func WithSampleRate(v int32) FilterOpt   { return WithKV("r", v) }
func WithLoop(v int) FilterOpt           { return WithKV("loop", v) }
func WithSize(v int64) FilterOpt         { return WithKV("size", v) }
func WithI(v float32) FilterOpt          { return WithKV("i", v) }
func WithLRA(v float32) FilterOpt        { return WithKV("lra", v) }
func WithTP(v float32) FilterOpt         { return WithKV("tp", v) }

func joinFilter(base string, opts ...FilterOpt) string {
	if len(opts) == 0 {
		return base
	}
	parts := []string{base}
	for _, opt := range opts {
		if opt == "" {
			continue
		}
		parts = append(parts, string(opt))
	}
	return strings.Join(parts, ":")
}
