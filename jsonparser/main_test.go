package main

import "testing"

func Test_parseJson(t *testing.T) {
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
			name:     `test case for parsJsonString`,
			input:    "\"hello world \"'",
			expected: `"hello world "'`,
			exec: func(input any) any {
				return parseJsonString(input.(string), &State{})
			},
		},
		{
			name:     `test case for parsJsonNumber`,
			input:    `1e+43, 443`,
			expected: 1e+43,
			exec: func(input any) any {
				return parseJsonNumber(input.(string), &State{})
			},
		},
		{
			name:     `test case for parsJsonBool`,
			input:    `   true,   false`,
			expected: true,
			exec: func(input any) any {
				return parseJsonBool(input.(string), &State{})
			},
		},
		{
			name:     `test case for parsJsonNull`,
			input:    `   null, `,
			expected: nil,
			exec: func(input any) any {
				return parseJsonNull(input.(string), &State{})
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
