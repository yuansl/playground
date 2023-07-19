package stream

import (
	"golang.org/x/exp/constraints"
)

type Stream[T, R any] struct {
	Value    []T
	parallel bool
}

func NewStream[T, R any](dataset []T) *Stream[T, R] {
	return &Stream[T, R]{Value: dataset}
}

func (s *Stream[T, R]) Filter(predict func(T) bool) *Stream[T, R] {
	values := make([]T, 0, len(s.Value))

	for _, v := range s.Value {
		if predict(v) {
			values = append(values, v)
		}
	}
	return NewStream[T, R](values)
}

func (s *Stream[T, R]) FindAny() *T {
	if len(s.Value) == 0 {
		return nil
	}

	return &s.Value[0]
}

func (s *Stream[T, R]) Map(fn func(T) T) *Stream[T, R] {
	outputs := make([]T, len(s.Value))
	for i := 0; i < len(s.Value); i++ {
		outputs[i] = fn(s.Value[i])
	}
	return NewStream[T, R](outputs)
}

func (s *Stream[T, R]) ForEach(fn func(T)) {
	for _, v := range s.Value {
		fn(v)
	}
}

func (s *Stream[T, R]) Limit(maxSize int) *Stream[T, R] {
	return NewStream[T, R](s.Value[:maxSize])
}

func (s *Stream[T, R]) Flatmap(mapper func(T) []T) *Stream[T, R] {
	result := make([]T, 0, len(s.Value))

	for _, v := range s.Value {
		result = append(result, mapper(v)...)
	}

	return NewStream[T, R](result)
}

func (s *Stream[T, R]) Count() int {
	return len(s.Value)
}

func (s *Stream[T, R]) Distinct() *Stream[T, R] {
	// TODO
	return s
}

func (s *Stream[T, R]) Min(less func(x, y T) bool) T {
	if len(s.Value) == 0 {
		return *(new(T))
	}

	min := s.Value[0]
	for i := 1; i < len(s.Value); i++ {
		if less(s.Value[i], min) {
			min = s.Value[i]
		}
	}
	return min
}

func (s *Stream[T, R]) Max(less func(x, y T) bool) T {
	if len(s.Value) == 0 {
		return *(new(T))
	}

	max := s.Value[0]
	for i := 1; i < len(s.Value); i++ {
		if less(max, s.Value[i]) {
			max = s.Value[i]
		}
	}
	return max
}

func (s *Stream[T, R]) Empty() *Stream[T, R] {
	return NewStream[T, R]([]T{})
}

func (s *Stream[T, R]) Reduce(identity T, fn func(v1, v2 T) T) T {
	if len(s.Value) == 0 {
		return identity
	}

	accum := identity

	for _, v := range s.Value {
		accum = fn(accum, v)
	}

	return accum
}

type Supplier[T any] interface {
	Get() T
}

type BiConsumer[T, R any] interface {
	Accept(T, R)
}

type BinaryOperator[T any] interface {
	Apply(T, T) T
}

type Function[T, R any] interface {
	Apply(T) R
}

type Collector[T, A, R any] struct {
	Supplier       func() A
	BiConsumer     func(A, T)
	BinaryOperator func(A, A) A
	Function       func(A) R
}

func (s *Stream[T, R]) Collect(collector Collector[T, any, R]) R {
	z := collector.Supplier()

	for _, v := range s.Value {
		collector.BiConsumer(z, v)
	}
	if s.parallel {
		// FIXME
		z = collector.BinaryOperator(z, z)
	}
	return collector.Function(z)
}

func GroupBy[T, R any, K comparable](inputs []T, getKey func(v T) K, fn2 func(K, []T) R) *Stream[R, T] {
	var result []R
	var groupBy = make(map[K][]T)

	for _, v := range inputs {
		key := getKey(v)
		groupBy[key] = append(groupBy[key], v)
	}
	for k, v := range groupBy {
		result = append(result, fn2(k, v))
	}
	return NewStream[R, T](result)
}

func AddArray[T constraints.Ordered](a, b []T) []T {
	if len(a) != len(b) {
		panic("BUG: len(a)!=len(b)")
	}

	c := make([]T, len(a))

	for i := 0; i < len(a); i++ {
		c[i] = a[i] + b[i]
	}
	return c
}
