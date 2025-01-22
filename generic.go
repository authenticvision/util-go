package util

func Map[T any, R any](list []T, fn func(T) R) (ret []R) {
	for _, v := range list {
		ret = append(ret, fn(v))
	}
	return
}

func Ref[T any](v T) *T {
	return &v
}
