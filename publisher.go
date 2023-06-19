package sinkv2

import (
	"context"
	"errors"
)

type Publisher interface {
	Publish(ctx context.Context, stat *DomainTrafficStat) error
}

type DomainTrafficStat struct {
	Domain  string
	Traffic int64
	Cdn     string
}

type domainsetPublisher struct {
}

func NewDomainTrafficPublisher() Publisher {
	return &domainsetPublisher{}
}

func (pub *domainsetPublisher) Publish(ctx context.Context, stat *DomainTrafficStat) error {
	return errors.New("not implemented")
}
