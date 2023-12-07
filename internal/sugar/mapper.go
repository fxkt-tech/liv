package sugar

type Single[I, O any] func(I) O

func Range[I, O any](inS []I, f Single[I, O]) []O {
	outs := make([]O, len(inS))
	for i, in := range inS {
		outs[i] = f(in)
	}
	return outs
}

func In[T int | string](elems []T, dest T) bool {
	for _, elem := range elems {
		if elem == dest {
			return true
		}
	}
	return false
}

func If[F func()](cond bool, f F) {
	if cond {
		f()
	}
}

func Filter[T any](slices []T, satisfied func(T) bool) []T {
	var results []T
	for _, s := range slices {
		if satisfied(s) {
			results = append(results, s)
		}
	}
	return results
}

func MapTo[T1, T2 any](slices []T1, deal func(T1) (T2, error)) []T2 {
	var results []T2
	for _, s := range slices {
		if t, err := deal(s); err == nil {
			results = append(results, t)
		}
	}
	return results
}
