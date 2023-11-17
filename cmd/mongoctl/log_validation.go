package main

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalid  = errors.New("mongodb: invalid argument")
	ErrDatabase = errors.New("mongodb: database error")
)

type DomainRetryFilter struct {
	DayHour time.Time `bson:"dayHour"`
}

type DomainRetryInfo struct {
	Domain       string
	DayHour      time.Time `bson:"dayHour"`
	CreateTime   time.Time `bson:"create_time"`
	RecheckState string    `bson:"recheck_state"`
	State        string    `bson:"state"`
}

type LogValidationRepository interface {
	DeleteRetryRecords(ctx context.Context, begin, end time.Time) error
	FindRetryRecords(ctx context.Context, begin, end time.Time) ([]DomainRetryInfo, error)
}

type logValidationRepository struct {
	domainRetry *DomainRetry
}

// DeleteRetryRecords implements LogValidationRepository.
func (r *logValidationRepository) DeleteRetryRecords(ctx context.Context, begin time.Time, end time.Time) error {
	for day := begin; day.Before(end); day = day.AddDate(0, 0, +1) {
		if err := r.domainRetry.Delete(ctx, &DomainRetryFilter{DayHour: begin}); err != nil {
			return err
		}
	}
	return nil
}

// FindRetryRecords implements LogValidationRepository.
func (r *logValidationRepository) FindRetryRecords(ctx context.Context, begin time.Time, end time.Time) ([]DomainRetryInfo, error) {
	var results []DomainRetryInfo

	for day := begin; day.Before(end); day = day.AddDate(0, 0, +1) {
		infos, err := r.domainRetry.Find(ctx, &DomainRetryFilter{DayHour: day})
		if err != nil {
			return nil, err
		}
		results = append(results, infos...)
	}
	return results, nil
}

var _ LogValidationRepository = (*logValidationRepository)(nil)

func NewLogValidationRepository(domainRetry *DomainRetry) LogValidationRepository {
	return &logValidationRepository{domainRetry: domainRetry}
}
