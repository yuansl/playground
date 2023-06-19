package main

import (
	"reflect"
	"testing"
)

func TestSet(t *testing.T) {
	// a test case consists of a @name, an @input, an @expected and
	// an @exec func

	cases := []struct {
		name     string
		disable  bool
		input    any
		expected any
		exec     func(input any) any
	}{
		{
			name: `test case for ...`,
			input: struct{ set1, set2, set3 []string }{
				set1: []string{"a", "b", "c"},
				set2: []string{"b", "a"},
				set3: []string{"c"},
			},
			expected: true,
			exec: func(input any) any {
				in := input.(struct{ set1, set2, set3 []string })
				set1 := newSet(in.set1)
				set2 := newSet(in.set2)
				return reflect.DeepEqual(setdiff(set1, set2), in.set3)
			},
		},
		{
			name: `test case for ...`,
			input: struct{ set1, set2, set3 []string }{
				set1: []string{"a", "b", "c"},
				set2: []string{"d", "e"},
				set3: []string{"a", "b", "c"},
			},
			expected: true,
			exec: func(input any) any {
				in := input.(struct{ set1, set2, set3 []string })
				set1 := newSet(in.set1)
				set2 := newSet(in.set2)
				return reflect.DeepEqual(setdiff(set1, set2), in.set3)
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
