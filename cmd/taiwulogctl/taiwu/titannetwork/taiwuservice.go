package titannetwork

import (
	"context"
	"time"

	"github.com/yuansl/playground/clients/titannetwork"
	"github.com/yuansl/playground/cmd/taiwulogctl/taiwu"
)

type taiwuService struct {
	*titannetwork.Client
}

// LogLink implements taiwu.TaiwuService.
func (srv *taiwuService) LogLink(ctx context.Context, domain string, timestamp time.Time) ([]taiwu.Link, error) {

	req := titannetwork.LogUrlRequest{Domain: domain, Timestamp: timestamp}
	switch srv.Version {
	case 2:
		res, err := srv.BossFlowLogUrlV2(ctx, &req)
		if err != nil {
			return nil, err
		}
		var links []taiwu.Link
		for _, it := range res.Urls {
			links = append(links, taiwu.Link{Url: it})
		}
		return links, nil
	case 1:
		fallthrough
	default:
		res, err := srv.BossFlowLogUrlV1(ctx, &req)
		if err != nil {
			return nil, err
		}
		return []taiwu.Link{{Url: res.Url}}, nil
	}
}

var _ taiwu.LogService = (*taiwuService)(nil)

func NewTaiwuService(c *titannetwork.Client) taiwu.LogService {
	return &taiwuService{Client: c}
}
