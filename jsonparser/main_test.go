package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

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
			input:    []byte(`"\"hello world \"'some"`),
			expected: `"hello world "'some`,
			exec: func(input any) any {
				return parseJsonString(input.([]byte), &State{})
			},
		},
		{
			name:     `test case for parsJsonString`,
			input:    []byte(`"lio"`),
			expected: `lio`,
			exec: func(input any) any {
				return parseJsonString(input.([]byte), &State{})
			},
		},
		{
			name:     `test case for parsJsonNumber`,
			input:    []byte(`1e+43, 443`),
			expected: 1e+43,
			exec: func(input any) any {
				return parseJsonNumber(input.([]byte), &State{})
			},
		},
		{
			name:     `test case for parsJsonBool`,
			input:    []byte(`   true,   false`),
			expected: true,
			exec: func(input any) any {
				return parseJsonBool(input.([]byte), &State{})
			},
		},
		{
			name:     `test case for parsJsonNull`,
			input:    []byte(`   null, `),
			expected: nil,
			exec: func(input any) any {
				return parseJsonNull(input.([]byte), &State{})
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

func Benchmark_parseJson(b *testing.B) {
	fp, err := os.Open(filepath.Join(os.Getenv("HOME"), ".cache/mintinstall/reviews.json"))
	if err != nil {
		fatal("os.Open error:", err)
	}
	defer fp.Close()

	data, err := io.ReadAll(io.LimitReader(fp, 30*MiB))
	if err != nil {
		fatal("io.ReadAll error:", err)
	}
	b.Run("parseJson", func(b *testing.B) {
		parseJson(data)
	})
	b.Run("json.Unmarshal", func(b *testing.B) {
		var v map[string]any

		json.Unmarshal(data, &v)
	})
}
