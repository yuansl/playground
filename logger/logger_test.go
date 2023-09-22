package logger

import (
	"context"
	"testing"

	"github.com/qbox/pili/base/qiniu/xlog.v1"
)

func TestLogger(t *testing.T) {
	// test case(s) consist of an @input value, an @expected value, and
	// an @exec func to get result
	cases := []struct {
		name     string
		disable  bool
		input    any
		expected any
		exec     func(input any) any
	}{
		{
			name:     "case for LoggerFromContext & ContextWithLogger",
			input:    nil,
			expected: nil,
			exec: func(input any) any {
				log := New()

				ctx := NewContext(context.Background(), log)

				func(ctx context.Context) {
					if FromContext(ctx) != log {
						t.Fail()
					}
				}(ctx)
				return nil
			},
		},
		{
			name:     "case for: Logger implements LoggerIdFromContext",
			input:    nil,
			expected: nil,
			exec: func(input any) any {
				log := New()

				ctx := NewContext(context.Background(), log)

				func(ctx context.Context) {
					reqId := log.(*xlog.Logger).ReqId()
					if IdFromContext(ctx) != reqId {
						t.Fail()
					}
				}(ctx)

				ctx = NewContext(context.Background(), noplogger)
				func(ctx context.Context) {
					if IdFromContext(ctx) != "" {
						t.Fail()
					}
				}(ctx)

				return nil
			},
		},
	}

	for _, c := range cases {
		if c.disable {
			continue
		}
		casex := c
		t.Run(casex.name, func(t *testing.T) {
			got := casex.exec(casex.input)
			if got != casex.expected {
				t.Fatalf("expected: `%v`, got: `%v`\n", casex.expected, got)
			}
		})
	}
}
