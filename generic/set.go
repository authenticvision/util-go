package generic

import "maps"

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](elems ...T) Set[T] {
	s := Set[T]{}
	for _, v := range elems {
		s[v] = struct{}{}
	}
	return s
}

func (s Set[T]) Contains(v T) bool {
	_, found := s[v]
	return found
}

func (s Set[T]) Subtract(o Set[T]) Set[T] {
	r := Set[T]{}
	for v := range s {
		if _, found := o[v]; !found {
			r[v] = struct{}{}
		}
	}
	return r
}

func (s Set[T]) Union(o Set[T]) Set[T] {
	r := maps.Clone(s)
	for v := range o {
		r[v] = struct{}{}
	}
	return r
}

func (s Set[T]) Intersect(o Set[T]) Set[T] {
	r := Set[T]{}
	for v := range s {
		if _, found := o[v]; found {
			r[v] = struct{}{}
		}
	}
	return r
}
