package stream

import (
	"testing"
	"time"
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

type Set[T comparable] struct {
	slots []T
}

func (set *Set[T]) Add(x T) {
	for _, v := range set.slots {
		if v == x {
			return
		}
	}
	set.slots = append(set.slots, x)
}

func (set *Set[T]) Get(index int) (t T) {
	if index > len(set.slots) {
		return
	}
	return set.slots[index]
}

func (set *Set[T]) Size() int {
	return len(set.slots)
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{}
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
						Supplier: func() any { return NewSet[int]() },
						BiConsumer: func(z any, x int) {
							z.(*Set[int]).Add(x)
						},
						Function: func(z any) []int {
							set := z.(*Set[int])
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
				type (
					Key struct {
						Domain string
						Day    time.Time
					}
					Value = []int
				)
				in := input.([]DomainCdnTraffic)
				out := NewStream[DomainCdnTraffic, []DomainTraffic](in).
					Filter(func(v DomainCdnTraffic) bool {
						return v.Cdn == "baidu"
					}).
					Map(func(v DomainCdnTraffic) DomainCdnTraffic {
						points := make([]int, 0, len(v.Points))

						for _, v := range v.Points {
							points = append(points, v*3)
						}
						return DomainCdnTraffic{Domain: v.Domain, Day: v.Day, Cdn: v.Cdn, Points: points}
					}).
					Collect(Collector[DomainCdnTraffic, any, []DomainTraffic]{
						Supplier: func() any { return NewSet[*DomainTraffic]() },
						BiConsumer: func(z any, x DomainCdnTraffic) {
							z.(*Set[*DomainTraffic]).
								Add(&DomainTraffic{Domain: x.Domain, Day: x.Day, Points: x.Points})
						},
						Function: func(z any) []DomainTraffic {
							set := z.(*Set[*DomainTraffic])
							result := make([]DomainTraffic, 0, set.Size())

							for i := 0; i < set.Size(); i++ {
								result = append(result, *set.Get(i))
							}
							return result
						},
					})
				t.Logf("out=%v\n", out)

				s2 := GroupBy[DomainCdnTraffic, DomainTraffic, Key](in,
					func(v DomainCdnTraffic) Key {
						return Key{Domain: v.Domain, Day: v.Day}
					},
					func(k Key, values []DomainCdnTraffic) DomainTraffic {
						points := []int{}

						for _, value := range values {
							if len(points) == 0 {
								points = make([]int, len(value.Points))
							}
							points = AddArray(points, value.Points)
						}
						return DomainTraffic{Domain: k.Domain, Day: k.Day, Points: points}
					},
				)
				t.Logf("s2 = %+v\n", s2)

				// for _, it := range out {
				// 	if v, ok := it.(GroupV[GroupKey, Value]); ok {
				// 		switch v.Key.Domain {
				// 		case "a":
				// 			if x := v.Values[0][0]; x != 12 {
				// 				t.Fatalf("mismatch: got %v, but expected 12", x)
				// 			}
				// 		case "b":
				// 			if x := v.Values[0][0]; x != 3 {
				// 				t.Fatalf("mismatch: got %v, but expected 3", x)
				// 			}
				// 		default:
				// 			panic("BUG: got unknown domain:" + v.Key.Domain)
				// 		}
				// 	}
				// }

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
