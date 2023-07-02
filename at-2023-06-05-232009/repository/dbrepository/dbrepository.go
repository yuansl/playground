package dbrepository

import (
	"context"
	"fmt"
	"time"

	"github.com/yuansl/playground/at-2023-06-05-232009/repository"
	"github.com/yuansl/playground/at-2023-06-05-232009/types"
)

var (
	beginOfTheProject = time.Date(2006, 1, 24, 0, 0, 0, 0, time.Local)
)

type Selector struct {
	Begin time.Time
	End   time.Time
}

type db interface { // CURD
	Save(ctx context.Context, s []types.Domain) error
	Find(ctx context.Context, selector Selector) ([]types.Domain, error)
}

type dbRepository struct {
	db
}

func (r *dbRepository) ListDomains(ctx context.Context) ([]types.Domain, error) {
	items, err := r.db.Find(ctx, Selector{Begin: beginOfTheProject})
	if err != nil {
		return nil, fmt.Errorf("db.Find: %v", err)
	}
	return items, nil
}

func NewRepository() repository.Repository {
	return &dbRepository{}
}
