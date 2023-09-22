package xsophon

import (
	"context"
	"fmt"
	"time"

	"github.com/yuansl/playground/clients/sophon"
	"github.com/yuansl/playground/cmd/bucketctl/bucket"
)

type bucketService struct {
	*sophon.Client
}

const _ACCOUNT_DEFAULT = "defy@qiniu.com"

// GetBucket implements bucket.BucketService.
func (srv *bucketService) GetBucket(ctx context.Context, domain string, time time.Time) (string, error) {
	r, err := srv.GetDomainDeliveryBucket(ctx, &sophon.BucketRequest{
		Domain:  domain,
		Since:   time,
		Account: _ACCOUNT_DEFAULT,
	})
	if err != nil {
		return "", fmt.Errorf("srv.GetDomainDeliveryBucket(domain=%s,time=%v): %v", domain, time, err)
	}
	return r.Bucket, nil
}

var _ bucket.BucketService = (*bucketService)(nil)

func NewBucketService(client *sophon.Client) bucket.BucketService {
	return &bucketService{Client: client}
}
