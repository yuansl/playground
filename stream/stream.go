package stream

import "fmt"

type Stream[T, R any] struct {
	value    []T
	parallel bool
}

func (s *Stream[T, R]) Filter(predict func(T) bool) *Stream[T, R] {
	values := make([]T, 0, len(s.value))

	for _, v := range s.value {
		if predict(v) {
			values = append(values, v)
		}
	}
	return NewStream[T, R](values)
}

func (s *Stream[T, R]) FindAny() *T {
	if len(s.value) == 0 {
		return nil
	}

	return &s.value[0]
}

func (s *Stream[T, R]) Map(fn func(T) T) *Stream[T, R] {
	outputs := make([]T, len(s.value))

	for i := 0; i < len(s.value); i++ {
		outputs[i] = fn(s.value[i])
	}
	return NewStream[T, R](outputs)
}

func (s *Stream[T, R]) ForEach(fn func(T) R) *Stream[R, T] {
	result := make([]R, 0, len(s.value))

	for _, v := range s.value {
		result = append(result, fn(v))
	}
	return NewStream[R, T](result)
}

func (s *Stream[T, R]) Limit(maxSize int) *Stream[T, R] {
	return NewStream[T, R](s.value[:maxSize])
}

func (s *Stream[T, R]) Flatmap(mapper func(T) []T) *Stream[T, R] {
	result := make([]T, 0, len(s.value))

	for _, v := range s.value {
		result = append(result, mapper(v)...)
	}

	return NewStream[T, R](result)
}

func (s *Stream[T, R]) Set() []T {
	return s.value
}

func (s *Stream[T, R]) Count() int {
	return len(s.value)
}

func (s *Stream[T, R]) Distinct(fn func(T) any) *Stream[T, R] {
	var uniq = make(map[any]T)
	var values []T

	for _, v := range s.value {
		key := fn(v)
		if _, exists := uniq[key]; exists {
			fmt.Printf("DEBUG: key: %s exists alreay for v: %+v\n", key, v)
		}
		uniq[key] = v
	}
	for _, v := range uniq {
		values = append(values, v)
	}
	return NewStream[T, R](values)
}

func (s *Stream[T, R]) Min(less func(x, y T) bool) T {
	if len(s.value) == 0 {
		return *(new(T))
	}

	min := s.value[0]
	for i := 1; i < len(s.value); i++ {
		if less(s.value[i], min) {
			min = s.value[i]
		}
	}
	return min
}

func (s *Stream[T, R]) Max(less func(x, y T) bool) T {
	if len(s.value) == 0 {
		return *(new(T))
	}

	max := s.value[0]
	for i := 1; i < len(s.value); i++ {
		if less(max, s.value[i]) {
			max = s.value[i]
		}
	}
	return max
}

func (s *Stream[T, R]) Empty() *Stream[T, R] {
	return NewStream[T, R]([]T{})
}

func (s *Stream[T, R]) Reduce(identity T, fn func(v1, v2 T) T) T {
	if len(s.value) == 0 {
		return identity
	}

	accum := identity

	for _, v := range s.value {
		accum = fn(accum, v)
	}

	return accum
}

func (s *Stream[R, T]) ReduceByKey(fn func(v1, v2 T) T) *Stream[R, T] {
	if len(s.value) == 0 {
		return NewStream[R, T]([]R{})
	}

	for _, v := range s.value {
		// FIXME
		_ = v
	}

	return NewStream[R, T](nil)
}

type Collector[T, A, R any] struct {
	Supplier       func() A
	BiConsumer     func(A, T)
	BinaryOperator func(A, A) A
	Function       func(A) R
}

func (s *Stream[T, R]) Collect(collector Collector[T, any, R]) R {
	z := collector.Supplier()

	for _, v := range s.value {
		collector.BiConsumer(z, v)
	}
	if s.parallel {
		// FIXME
		z = collector.BinaryOperator(z, z)
	}
	return collector.Function(z)
}

func NewStream[T, R any](dataset []T) *Stream[T, R] {
	return &Stream[T, R]{value: dataset}
}

// T is the type of inputs
// R is the type of Outputs
// K is the group by key
func GroupBy[T, R any, K comparable](inputs []T, getKey func(v T) K, convert func(K, []T) R) *Stream[R, T] {
	var result []R
	var groupBy = make(map[K][]T)

	for _, v := range inputs {
		key := getKey(v)
		groupBy[key] = append(groupBy[key], v)
	}
	for k, v := range groupBy {
		result = append(result, convert(k, v))
	}
	return NewStream[R, T](result)
}

func GroupBy2[T any, R struct {
	Key    K
	Values []T
}, K comparable](inputs []T, getKey func(v T) K) *Stream[R, T] {
	var groupBy = make(map[K][]T)
	var result []R

	for _, v := range inputs {
		key := getKey(v)
		groupBy[key] = append(groupBy[key], v)
	}
	for k, v := range groupBy {
		result = append(result, R{Key: k, Values: v})
	}
	return NewStream[R, T](result)
}
