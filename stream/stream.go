package stream

type Stream[T any, K comparable, R any] struct {
	value    []T
	parallel bool
}

type Predicate[T any] interface {
	Test(T) bool
}

type PredicateFn[T any] func(T) bool

func (f PredicateFn[T]) Test(t T) bool {
	return f(t)
}

func (s *Stream[T, K, R]) Filter(predict Predicate[T]) *Stream[T, K, R] {
	values := make([]T, 0, len(s.value))

	for _, v := range s.value {
		if predict.Test(v) {
			values = append(values, v)
		}
	}
	return NewStream[T, K, R](values)
}

func (s *Stream[T, K, R]) FindAny() *T {
	if len(s.value) == 0 {
		return nil
	}

	return &s.value[0]
}

type Function[T, R any] interface {
	Apply(T) R
}

type FunctionFn[T, R any] func(T) R

func (f FunctionFn[T, R]) Apply(t T) R {
	return f(t)
}

func (s *Stream[T, K, R]) Map(fn Function[T, R]) *Stream[R, K, any] {
	outputs := make([]R, len(s.value))

	for i := 0; i < len(s.value); i++ {
		outputs[i] = fn.Apply(s.value[i])
	}
	return NewStream[R, K, any](outputs)
}

type Consumer[T any] interface {
	Accept(T)
}

type ConsumerFn[T any] func(T)

func (f ConsumerFn[T]) Accept(t T) {
	f(t)
}

func (s *Stream[T, K, R]) ForEach(fn Consumer[*T]) *Stream[T, K, R] {
	result := make([]T, len(s.value))

	copy(result, s.value)

	for i := 0; i < len(result); i++ {
		v := &result[i]
		fn.Accept(v)
	}
	return NewStream[T, K, R](result)
}

func (s *Stream[T, K, R]) Limit(maxSize int) *Stream[T, K, R] {
	return NewStream[T, K, R](s.value[:maxSize])
}

func (s *Stream[T, K, R]) Flatmap(mapper Function[T, *Stream[R, K, T]]) *Stream[R, K, T] {
	result := make([]R, 0, len(s.value))

	for _, v := range s.value {
		result = append(result, mapper.Apply(v).Collect()...)
	}

	return NewStream[R, K, T](result)
}

func (s *Stream[T, K, R]) Collect() []T {
	return s.value
}

func (s *Stream[T, K, R]) Count() int {
	return len(s.value)
}

func (s *Stream[T, K, R]) Distinct() *Stream[T, K, R] {
	var uniq = make(map[any]struct{})
	var values []T

	for _, v := range s.value {
		uniq[v] = struct{}{}
	}
	for v := range uniq {
		values = append(values, v.(T))
	}
	return NewStream[T, K, R](values)
}

func (s *Stream[T, K, R]) Min(less func(x, y T) bool) T {
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

func (s *Stream[T, K, R]) Max(less func(x, y T) bool) T {
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

func (s *Stream[T, K, R]) Empty() bool {
	return len(s.value) == 0
}

type Tuple[K comparable, V any] struct {
	Key   K
	Value V
}

// GroupBy returns []{K;[]T} for a dataset []T
func (s *Stream[T, K, V]) GroupBy(keyOf func(T) K) *Stream[Tuple[K, any], K, V] {
	type Value = Tuple[K, any]
	var values []Value
	var groupBy = make(map[K][]T)

	for _, v := range s.value {
		key := keyOf(v)
		groupBy[key] = append(groupBy[key], v)
	}
	for k, v := range groupBy {
		values = append(values, Value{Key: k, Value: v})
	}
	return NewStream[Value, K, V](values)
}

type BinaryOperator[T any] interface {
	Apply(T, T) T
}

type BinaryOpFn[T any] func(T, T) T

func (f BinaryOpFn[T]) Apply(v1, v2 T) T { return f(v1, v2) }

// ReduceByKey reduces a dataset []T, which ([]{K; []V}) created by GroupBy function
// returns []{K;V}
func (s *Stream[T, K, V]) ReduceByKey(key func(T) Tuple[K, []V], reducef BinaryOperator[V]) *Stream[Tuple[K, any], K, V] {
	type Value = Tuple[K, any]
	var values []Value

	for _, v := range s.value {
		gv := key(v)
		accum := gv.Value[0]

		for i := 1; i < len(gv.Value); i++ {
			accum = reducef.Apply(accum, gv.Value[i])
		}
		values = append(values, Value{Key: gv.Key, Value: accum})
	}
	return NewStream[Value, K, V](values)
}

func (s *Stream[T, K, R]) Reduce(identity T, fn func(v1, v2 T) T) T {
	if len(s.value) == 0 {
		return identity
	}

	accum := identity

	for _, v := range s.value {
		accum = fn(accum, v)
	}

	return accum
}

// T -> (A)intermediate -> R
type Aggregator[T, A, R any] struct {
	Supplier       func() A
	Transform      func(A, T)
	BinaryOperator func(A, A) A
	Collect        func(A) []R
}

// T - type of the input
// R - type of the result
func (s *Stream[T, K, R]) Aggregate(aggregator Aggregator[T, any, R]) []R {
	z := aggregator.Supplier()

	for _, v := range s.value {
		aggregator.Transform(z, v)
	}
	if s.parallel {
		panic("TODO: not implemented")
		// z = collector.BinaryOperator(z, z)
	}
	return aggregator.Collect(z)
}

func NewStream[T any, K comparable, R any](dataset []T) *Stream[T, K, R] {
	return &Stream[T, K, R]{value: dataset}
}
