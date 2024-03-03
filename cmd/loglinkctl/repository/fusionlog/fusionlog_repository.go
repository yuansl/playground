package fusionlog

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/util"
	"github.com/yuansl/playground/cmd/loglinkctl/repository"
)

var ErrInvalid = errors.New("repository: invalid argument")

type Filter struct {
	ID        string
	Domain    string
	Url       string
	HourBegin time.Time
	HourEnd   time.Time
	Business  string
	Limit     int
}
type LoglinkStorage interface {
	Query(ctx context.Context, filter *Filter) ([]LogLink, error)
	Update(ctx context.Context, link *LogLink) error
	Delete(ctx context.Context, link *LogLink) error
}

type fusionlogRepository struct {
	fusionlog LoglinkStorage
}

type (
	LogLink     = repository.LogLink
	LinkOptions = repository.LinkOptions
)

// GetLink implements LinkRepository.
func (repo *fusionlogRepository) GetLinks(ctx context.Context, begin time.Time, end time.Time, opts ...*LinkOptions) ([]LogLink, error) {
	var options LinkOptions
	var links []LogLink
	var err error

	if len(opts) > 0 {
		options = *opts[0]
	}
	err = util.WithRetry(ctx, func() error {
		links, err = repo.fusionlog.Query(ctx, &Filter{
			Domain:    options.Domain,
			HourBegin: begin,
			HourEnd:   end,
			Business:  options.Business.String(),
			Limit:     10000,
		})
		return err
	})
	if err != nil {
		util.Fatal("fusionlog.Query: %v", err)
	}
	llen := len(links)
	if options.Filter != nil {
		links = slices.DeleteFunc(links, func(link LogLink) bool { return !options.Filter(&link) })
	}

	logger.FromContext(ctx).Infof("fetched %d records after filter, and %d effective records in total\n", llen, len(links))

	return links, nil
}

func (repo *fusionlogRepository) DeleteLinks(ctx context.Context, links ...LogLink) error {
	for _, link := range links {
		if err := repo.fusionlog.Delete(ctx, &link); err != nil {
			return fmt.Errorf("repo.fusionlog.Delete: %w", err)
		}
	}
	return nil
}

// SetDownloadUrl implements LinkRepository.
func (repo *fusionlogRepository) SetDownloadUrl(ctx context.Context, link *LogLink, url string) error {
	logger.FromContext(ctx).Infof("link: %+v, new url: %q\n", link, url)

	link.Url = url

	return repo.fusionlog.Update(ctx, link)
}

func NewLinkRepository(fusionlog LoglinkStorage) (repository.LinkRepository, error) {
	return &fusionlogRepository{
		fusionlog: fusionlog,
	}, nil
}
