package main

import (
	"context"

	"github.com/yuansl/playground/cmd/taiwulogctl/sinker"
)

type nopTrafficSinker struct{}

// Sink implements TrafficSinker.
func (*nopTrafficSinker) Sink(ctx context.Context, _ []sinker.TrafficStat) error {
	return nil // do nothing
}

var _ sinker.TrafficSinker = (*nopTrafficSinker)(nil)

func NewNopTrafficSinker() sinker.TrafficSinker {
	return &nopTrafficSinker{}
}
