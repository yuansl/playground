package etlservice

import (
	"context"
	"errors"
	"time"
)

var ErrInvalid = errors.New("etlservice: Invalid argument")

type ETLService interface {
	EtlRetry(ctx context.Context, domains []string, cdn string, start, end time.Time) error
	Sync(ctx context.Context, domains []string, start, end time.Time) error
}
