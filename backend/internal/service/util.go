package service

func setToSlice[T comparable](m map[T]struct{}) []T {
	out := make([]T, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

