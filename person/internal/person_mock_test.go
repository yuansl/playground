package internal

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestPersonMock(t *testing.T) {
	// a test case consists of a @name, an @input, an @expected and
	// an @exec func

	debug := func(args ...interface{}) {
		_, file, line, _ := runtime.Caller(-1)
		fmt.Printf("%s:%d: %v", file, line, args)
	}

	debug("debug will print source code location of this line")

	cases := []struct {
		name     string
		disable  bool
		input    interface{}
		expected interface{}
		exec     func(input interface{}) interface{}
	}{
		{
			name:     "test case for ...",
			input:    nil,
			expected: "Smith",
			exec: func(input interface{}) interface{} {
				mockctrl := gomock.NewController(t)

				mock := NewMockPerson(mockctrl)
				mock.
					EXPECT().
					Name().
					Return("Smith").
					AnyTimes()

				return mock.Name()
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
