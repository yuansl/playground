package main

import (
	"context"
	"testing"
	"time"
)

func TestAddDate(t *testing.T) {
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
				r := mockTrafficRepository{
					timeseries: Timesereis{1, 2, 3, 4, 5, 6, 7, 1200},
				}
				ctx := context.TODO()
				cdn := "volcengine2"
				from := time.Date(2023, 3, 1, 0, 0, 0, 0, time.Local)
				to := time.Date(2023, 3, 2, 0, 0, 0, 0, time.Local)

				peak95 := statPeak95(ctx, r, cdn, from, to)
				if peak95 != 1200 {
					t.Fatalf("mismatch: got %d, expected 1204\n", peak95)
				}
				avrpeak := statAveragePeakOf(ctx, r, cdn, from, to)
				if avrpeak != 32 {
					t.Fatalf("mismatch: got %f, expected 32\n", avrpeak)
				}
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

type mockTrafficRepository struct {
	timeseries Timesereis
}

func (x mockTrafficRepository) GetTimeseriesOf(ctx context.Context, cdn string, from time.Time, to time.Time, domains ...string) (Timesereis, error) {
	return x.timeseries, nil
}
