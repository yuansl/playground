package repository

import (
	"context"
	"playground/at-2023-06-05-232009/types"
)

type Repository interface {
	ListDomains(ctx context.Context) ([]types.Domain, error)
}
