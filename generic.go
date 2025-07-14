package util

func Map[T any, R any](list []T, fn func(T) R) (ret []R) {
	for _, v := range list {
		ret = append(ret, fn(v))
	}
	return
}

func Filter[T any](list []T, fn func(T) bool) (ret []T) {
	for _, v := range list {
		if fn(v) {
			ret = append(ret, v)
		}
	}
	return
}

func FilterMap[T any, R any](list []T, fn func(T) (R, bool)) (ret []R) {
	for _, v := range list {
		if r, ok := fn(v); ok {
			ret = append(ret, r)
		}
	}
	return
}

func Unique[T comparable](list []T) (ret []T) {
	m := make(map[T]struct{})
	for _, v := range list {
		m[v] = struct{}{}
	}
	return Keys(m)
}

func Ref[T any](v T) *T {
	return &v
}

func Keys[T comparable, X any](m map[T]X) []T {
	var keys []T
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
