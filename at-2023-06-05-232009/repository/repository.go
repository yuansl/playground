package repository

import (
	"context"

	"github.com/yuansl/playground/at-2023-06-05-232009/types"
)

type Repository interface {
	ListDomains(ctx context.Context) ([]types.Domain, error)
}
