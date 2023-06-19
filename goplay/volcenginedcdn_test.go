package fscdn

import (
	"context"
	"testing"
	"time"

	"github.com/qbox/net-deftones/logger"
)

func TestVolcengineDcdn(t *testing.T) {
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
				svc, err := NewVolcengineDcdnService(&Config{
					Cdn:   "volcenginedcdn",
					CdnAk: "AKLTNDg0MGE4NjQyMjEyNGU0M2I2ZWJlZDEzY2E4ZTk0ZTY",
					CdnSk: "TVRBME1XRmlORFF6TVdFM05HUmtZemxqTVdKbFpUSXpOelF6WmpVM05UWQ==",
				})
				if err != nil {
					t.Fatal(err)
				}
				ctx := logger.ContextWithLogger(context.Background(), logger.New())

				start := time.Date(2022, 6, 16, 0, 0, 0, 0, time.Local)
				end := start.Add(24 * time.Hour)
				domains := []CdnDomain{{Domain: "testprd798.qnqcdn.net", Key: "testprd798.qnqcdn.net"}}

				bw, err := svc.FetchCdnDomainsBandwidth(ctx, domains, start, end)
				if err != nil {
					t.Fatal(err)
				}
				t.Logf("bw: %+v\n", bw)

				qps, err := svc.FetchCdnDomainsQPS(ctx, domains, start, end)
				if err != nil {
					t.Fatal(err)
				}
				t.Logf("qps: %+v\n", qps)

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
