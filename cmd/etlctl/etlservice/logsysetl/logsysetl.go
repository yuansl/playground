package logsysetl

import (
	"context"
	"fmt"
	"time"

	"github.com/yuansl/playground/clients/logetl"
	"github.com/yuansl/playground/cmd/etlctl/etlservice"
	"github.com/yuansl/playground/util"
)

const _SYNC_BATCHSIZE_MAX = 200

type logetlService struct {
	*logetl.Client
	options *etlOptions
}

type etlOptions struct {
	syncBatchSize int
}

var _ etlservice.ETLService = (*logetlService)(nil)

func (srv *logetlService) Sync(ctx context.Context, domains []string, start, end time.Time) error {
	for i := 0; i < srv.options.syncBatchSize; i += srv.options.syncBatchSize {
		i0, i1 := i, i+srv.options.syncBatchSize
		if i1 > len(domains) {
			i1 = len(domains)
		}
		res, err := srv.GetUnifyDaySyncs(ctx, &logetl.DaySyncsRequest{
			Domains: domains[i0:i1],
			Start:   start,
			End:     end,
		})
		if err != nil {
			return fmt.Errorf("srv.SendDaySyncsRequest: %v => '%v'", err, res)
		}
	}
	return nil
}

func (srv *logetlService) EtlRetry(ctx context.Context, domains []string, cdn string, start, end time.Time) error {
	if start.IsZero() || end.IsZero() {
		return fmt.Errorf("%w: start or end must not be zero", etlservice.ErrInvalid)
	}
	if res, err := srv.SendEtlRetryRequest(ctx, &logetl.EtlRetryRequest{
		Cdn:          cdn,
		Domains:      domains,
		Start:        start,
		End:          end,
		Force:        true,
		Manual:       true,
		IgnoreOnline: true,
	}); err != nil {
		return fmt.Errorf("srv.EtlRetryRequest: %v => '%+v'", err, res)
	}
	return nil
}

type EtlOption util.Option

func WithSyncBatchSize(size int) EtlOption {
	return util.OptionFunc(func(op any) {
		op.(*etlOptions).syncBatchSize = size
	})
}

func NewETLService(cli *logetl.Client, opts ...EtlOption) etlservice.ETLService {
	var options etlOptions

	for _, op := range opts {
		op.Apply(&options)
	}
	if options.syncBatchSize <= 0 || options.syncBatchSize > _SYNC_BATCHSIZE_MAX {
		options.syncBatchSize = _SYNC_BATCHSIZE_MAX
	}
	return &logetlService{Client: cli, options: &options}
}
