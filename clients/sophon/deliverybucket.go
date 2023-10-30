package sophon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/yuansl/playground/util"
)

const _FUSION_SUFY_ENDPOINT = "http://fusiondomainproxy.qcdn-sophon-sufy.qa.qiniu.io"

var ErrInvalid = errors.New("clients.sophon: invalid argument")

type Client struct {
	*http.Client
}

type BucketRequest struct {
	Domain  string
	Since   time.Time
	Account string
}

type BucketResponse struct {
	Bucket string
}

func (client *Client) GetDomainDeliveryBucket(ctx context.Context, req *BucketRequest) (*BucketResponse, error) {
	var res BucketResponse
	var extraHeader = make(http.Header)
	var query = make(url.Values)

	extraHeader.Set("Account", req.Account)
	query.Add("domain", req.Domain)
	if !req.Since.IsZero() {
		query.Add("time", strconv.Itoa(int(req.Since.Unix())))
	}
	if err := client.send(ctx,
		&httpRequest{
			url:          _FUSION_SUFY_ENDPOINT + "/admin/domain/deliverybucket",
			method:       http.MethodGet,
			query:        query,
			extraHeaders: extraHeader,
		}, &res); err != nil {
		return nil, fmt.Errorf("client.sendRequest: %v", err)
	}
	return &res, nil
}

type ClientOption util.Option

type clientOptions struct {
	accessKey string
	secretKey string
}

func WithCredentials(accessKey, secretKey string) ClientOption {
	return util.OptionFunc(func(op any) {
		op.(*clientOptions).accessKey = accessKey
		op.(*clientOptions).secretKey = secretKey
	})
}

func NewClient(opts ...ClientOption) *Client {
	var transport = http.DefaultTransport
	var options clientOptions

	for _, op := range opts {
		op.Apply(&options)
	}
	if options.accessKey != "" && options.secretKey != "" {
		transport = util.NewAuthorizedTransport(options.accessKey, options.secretKey)
	}
	return &Client{
		Client: &http.Client{
			Transport: transport,
		},
	}
}
