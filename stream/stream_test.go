package stream

import (
	"cmp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type DomainCdnTraffic struct {
	Domain string
	Day    time.Time
	Cdn    string
	Points []int
}

type DomainTraffic struct {
	Domain string
	Day    time.Time
	Points []int
}

func AddArray[T cmp.Ordered](a, b []T) []T {
	if len(a) != len(b) {
		panic("BUG: len(a)!=len(b)")
	}

	c := make([]T, len(a))

	for i := 0; i < len(a); i++ {
		c[i] = a[i] + b[i]
	}
	return c
}

func TestStream(t *testing.T) {
	// a test case consists of a @name, an @input, an @expected and
	// an @exec func
	now := time.Now()

	cases := []struct {
		name     string
		disable  bool
		input    any
		expected any
		exec     func(input any) any
	}{
		{
			name:     `test case for [1,2,3,3,4,5]`,
			input:    []int{1, 2, 3, 3, 4, 5},
			expected: nil,
			exec: func(input any) any {
				in := input.([]int)
				out := NewStream[int, int, int](in).
					Map(FunctionFn[int, int](func(v int) int {
						return v * 3
					})).
					Flatmap(FunctionFn[int, *Stream[any, int, int]](func(v int) *Stream[any, int, int] {
						return NewStream[any, int, int]([]any{v, v + 1})
					})).
					Map(FunctionFn[any, int](
						func(v any) int {
							return v.(int)
						})).
					Distinct().
					Collect()
				if len(out) != 10 {
					t.Fatalf("mismatch: got %d items, but expected 10", len(out))
				}
				return nil
			},
		},
		{
			name: `test case for `,
			input: []DomainCdnTraffic{
				{Domain: "a", Day: now, Cdn: "aliyun", Points: []int{1, 2, 3}},
				{Domain: "a", Day: now, Cdn: "baidu", Points: []int{4, 5, 6}},
				{Domain: "a", Day: now, Cdn: "baishanyun", Points: []int{7, 8, 9}},
				{Domain: "b", Day: now, Cdn: "baishanyun", Points: []int{10, 11, 12}},
				{Domain: "b", Day: now, Cdn: "baidu", Points: []int{1, 2, 3}},
			},
			expected: nil,
			exec: func(input any) any {
				in := input.([]DomainCdnTraffic)
				out := NewStream[DomainCdnTraffic, any, DomainTraffic](in).
					Filter(PredicateFn[DomainCdnTraffic](func(v DomainCdnTraffic) bool {
						return v.Cdn == "baidu"
					})).
					ForEach(ConsumerFn[*DomainCdnTraffic](func(v *DomainCdnTraffic) {
						for i := 0; i < len(v.Points); i++ {
							p := &v.Points[i]
							*p *= 3
						}
					})).
					Map(FunctionFn[DomainCdnTraffic, DomainTraffic](
						func(x DomainCdnTraffic) DomainTraffic {
							return DomainTraffic{Domain: x.Domain, Day: x.Day, Points: x.Points}
						})).
					Collect()

				assert.Equal(t, 2, len(out))
				if out[0].Domain == "a" {
					assert.Equal(t, []int{12, 15, 18}, out[0].Points)
					assert.Equal(t, []int{3, 6, 9}, out[1].Points)
				} else if out[0].Domain == "b" {
					assert.Equal(t, []int{3, 6, 9}, out[0].Points)
					assert.Equal(t, []int{12, 15, 18}, out[1].Points)
				} else {
					t.Fatal("mismatch: out[0].domain must be one of a or b")
				}

				type GroupByKey struct {
					Domain string
					Day    time.Time
				}

				out2 := NewStream[DomainCdnTraffic, GroupByKey, DomainCdnTraffic](in).
					GroupBy(func(t DomainCdnTraffic) GroupByKey {
						return GroupByKey{Domain: t.Domain, Day: t.Day}
					}).
					ReduceByKey(
						func(gv Tuple[GroupByKey, any]) Tuple[GroupByKey, []DomainCdnTraffic] {
							return Tuple[GroupByKey, []DomainCdnTraffic]{
								Key: gv.Key, Value: gv.Value.([]DomainCdnTraffic),
							}
						},
						BinaryOpFn[DomainCdnTraffic](
							func(v1, v2 DomainCdnTraffic) DomainCdnTraffic {
								_v1 := v1
								_v1.Points = AddArray(_v1.Points, v2.Points)
								return _v1
							})).
					Map(FunctionFn[Tuple[GroupByKey, any], DomainCdnTraffic](
						func(gv Tuple[GroupByKey, any]) DomainCdnTraffic {
							v := gv.Value.(DomainCdnTraffic)
							return DomainCdnTraffic{Domain: v.Domain, Day: v.Day, Points: v.Points}
						})).
					Collect()

				t.Logf("s3=%+v\n", out2)

				return nil
			},
		},
	}

	for _, c := range cases {
		if c.disable {
			continue
		}

		test := c
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.exec(test.input)
			if got != test.expected {
				t.Fatalf("expected: `%v`, got: `%v`\n", test.expected, got)
			}
		})
	}
}
