package main

import (
	"encoding/json"
	"testing"
)

func Test_json_Unmarshal(t *testing.T) {
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
			name:     `test case for ...`,
			input:    nil,
			expected: nil,
			exec: func(input any) any {
				var v any

				if err := json.Unmarshal([]byte(`{"hello":"world"}`), &v); err != nil {
					t.Fatal(err)
				}
				t.Logf("v = %v\n", v)
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
