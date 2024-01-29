package titannetwork

import (
	"context"
	"time"

	"github.com/qbox/net-deftones/util"
	"github.com/yuansl/playground/clients/titannetwork"
	"github.com/yuansl/playground/cmd/taiwulogctl/logservice"
)

type taiwuService struct {
	client     *titannetwork.Client
	apiVersion int
}

const (
	APIVERSION_2 = 2
	APIVERSION_1 = 1
)

// LogLink implements taiwu.TaiwuService.
func (srv *taiwuService) LogLinks(ctx context.Context, domain string, timestamp time.Time, token ...string) ([]logservice.Link, error) {
	req := titannetwork.LogUrlRequest{
		Domain:    domain,
		Timestamp: timestamp,
	}
	if len(token) > 0 {
		req.Token = token[0]
	}
	switch srv.apiVersion {
	case APIVERSION_2:
		res, err := srv.client.BossFlowLogUrlV2(ctx, &req)
		if err != nil {
			return nil, err
		}
		var links []logservice.Link
		for _, it := range res.Urls {
			links = append(links, logservice.Link{Url: it})
		}
		return links, nil
	case APIVERSION_1:
		fallthrough
	default:
		res, err := srv.client.BossFlowLogUrlV1(ctx, &req)
		if err != nil {
			return nil, err
		}
		return []logservice.Link{{Url: res.Url}}, nil
	}
}

var _ logservice.LogService = (*taiwuService)(nil)

type Option util.Option

type clientOptions struct{ version int }

func WithVersion(version int) Option {
	return util.OptionFunc(func(opt any) {
		opt.(*clientOptions).version = version
	})
}

func NewTaiwuLogService(c *titannetwork.Client, opts ...Option) logservice.LogService {
	var options clientOptions
	for _, opt := range opts {
		opt.Apply(&options)
	}
	if options.version == 0 {
		options.version = APIVERSION_2
	}
	return &taiwuService{client: c, apiVersion: options.version}
}
