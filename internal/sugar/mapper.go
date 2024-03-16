package sugar

type Single[I, O any] func(I) O

func Multi[I, O any](inS []I, f Single[I, O]) []O {
	outs := make([]O, len(inS))
	for i, in := range inS {
		outs[i] = f(in)
	}
	return outs
}

func In[T comparable](elems []T, dest T) bool {
	for _, elem := range elems {
		if elem == dest {
			return true
		}
	}
	return false
}

// 为了能将if一行写下而存在，适用于极简场景，其它情况下不要使用这个函数！
func If[F func()](cond bool, f F) {
	if cond {
		f()
	}
}

// 为了能将for range一行写下而存在，适用于极简场景，其它情况下不要使用这个函数！
func Range[T any](elems []T, f func(T)) {
	for _, elem := range elems {
		f(elem)
	}
}

// 过滤列表
func Filter[T any](slices []T, satisfied func(T) bool) []T {
	var results []T
	for _, s := range slices {
		if satisfied(s) {
			results = append(results, s)
		}
	}
	return results
}

// 将一个类型的列表转换成另一个类型的列表
func MapTo[T1, T2 any](slices []T1, deal func(T1) (T2, error)) []T2 {
	var results []T2
	for _, s := range slices {
		if t, err := deal(s); err == nil {
			results = append(results, t)
		}
	}
	return results
}

func Get[T comparable](value, dv T) T {
	var zero T
	if value != zero {
		return value
	}
	return dv
}
