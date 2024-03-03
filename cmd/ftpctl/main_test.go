package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_aligned2time(t *testing.T) {
	timestamp := time.Date(2024, 2, 1, 1, 36, 48, 0, time.Local)
	aligned := time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), 0, 0, 0, 0, timestamp.Location())
	cases := []struct{ t, aligned time.Time }{{timestamp, aligned}}

	for _, it := range cases {
		t2 := alignoftime(it.t, 24*time.Hour)

		assert.Equal(t, t2.Unix(), it.aligned.Unix())
	}
}
