package titannetwork

import (
	"context"
	"testing"
	"time"

	"github.com/yuansl/playground/clients/titannetwork"
)

func alignedtime(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, t.Hour(), t.Minute()/5*5, 0, 0, t.Location())
}

func Test_taiwuService(t *testing.T) {
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
				client := titannetwork.NewClient(
					titannetwork.WithCredential("qiniu", []byte("a5c90e5370c80067a2ac78aab1badb90")),
				)
				srv := NewTaiwuLogService(client)

				links, err := srv.LogLinks(context.TODO(), "audiosdk.xmcdn.com", alignedtime(time.Now().Add(-5*time.Hour)), "386BD183")
				if err != nil {
					t.Fatal(err)
				}
				t.Logf("links=%+v\n", links)
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
