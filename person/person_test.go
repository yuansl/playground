package person

import "testing"

func TestPerson(t *testing.T) {
	// a test case consists of a @name, an @input, an @expected and
	// an @exec func

	cases := []struct {
		name     string
		disable  bool
		input    interface{}
		expected interface{}
		exec     func(input interface{}) interface{}
	}{
		{
			name:     "test case for ...",
			input:    -9,
			expected: nil,
			exec: func(input interface{}) interface{} {
				person := NewPerson()

				t.Logf("-9/5=%d, -9%%5=%d, person=%v\n", -9/5, -9%5, person)
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
			got := test.exec(test.input)
			if got != test.expected {
				t.Fatalf("expected: `%v`, got: `%v`\n", test.expected, got)
			}
		})
	}
}
