package generic

func AnySlice[T any](in []T) []any {
	return Map(in, func(t T) any {
		return t
	})
}

func Map[T, U any](in []T, f func(T) U) []U {
	result := make([]U, len(in))
	for i := range in {
		result[i] = f(in[i])
	}
	return result
}
