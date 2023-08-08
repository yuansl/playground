package stream

import (
	"testing"
	"time"

	"github.com/yuansl/playground/utils"
	"golang.org/x/exp/constraints"
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
				out := NewStream[int, []int](in).
					Map(func(v int) int {
						return v * 3
					}).
					Flatmap(func(v int) []int {
						return []int{v, v + 1}
					}).
					Collect(Collector[int, any, []int]{
						Supplier: func() any { return utils.NewSet[int]() },
						BiConsumer: func(z any, x int) {
							z.(*utils.Set[int]).Add(x)
						},
						Function: func(z any) []int {
							set := z.(*utils.Set[int])
							result := make([]int, 0, set.Size())
							for i := 0; i < set.Size(); i++ {
								result = append(result, set.Get(i))
							}
							return result
						},
					})
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
				out := NewStream[DomainCdnTraffic, []DomainTraffic](in).
					Filter(func(v DomainCdnTraffic) bool {
						return v.Cdn == "baidu"
					}).
					Map(func(v DomainCdnTraffic) DomainCdnTraffic {
						for i := 0; i < len(v.Points); i++ {
							p := &v.Points[i]

							*p *= 3
						}
						return v
					}).
					Collect(Collector[DomainCdnTraffic, any, []DomainTraffic]{
						Supplier: func() any { return utils.NewSet[*DomainTraffic]() },
						BiConsumer: func(z any, x DomainCdnTraffic) {
							z.(*utils.Set[*DomainTraffic]).
								Add(&DomainTraffic{Domain: x.Domain, Day: x.Day, Points: x.Points})
						},
						Function: func(z any) []DomainTraffic {
							set := z.(*utils.Set[*DomainTraffic])
							result := make([]DomainTraffic, 0, set.Size())

							for i := 0; i < set.Size(); i++ {
								result = append(result, *set.Get(i))
							}
							return result
						},
					})
				t.Logf("out=%v\n", out)

				type GroupByKey struct {
					Domain string
					Day    time.Time
				}

				s2 := GroupBy[DomainCdnTraffic, DomainTraffic, GroupByKey](in,
					func(v DomainCdnTraffic) GroupByKey {
						return GroupByKey{Domain: v.Domain, Day: v.Day}
					},
					func(k GroupByKey, values []DomainCdnTraffic) DomainTraffic {
						points := []int{}

						for _, value := range values {
							if len(points) == 0 {
								points = make([]int, len(value.Points))
							}
							points = AddArray(points, value.Points)
						}
						return DomainTraffic{Domain: k.Domain, Day: k.Day, Points: points}
					},
				).Set()
				t.Logf("s2 = %+v\n", s2)

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
