package sophon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	"github.com/yuansl/playground/utils"
)

const _FUSION_SUFY_ENDPOINT = "http://fusiondomainproxy.qcdn-sophon-sufy.qa.qiniu.io"

var ErrInvalid = errors.New("clients.sophon: invalid argument")

type Client struct {
	*http.Client
}

type httpRequest struct {
	url          string
	method       string
	query        url.Values
	body         io.Reader
	contentType  string
	extraHeaders http.Header
}

func (client *Client) send(ctx context.Context, req *httpRequest, res any) error {
	if req.method == "" {
		return fmt.Errorf("%w: method must not be empty", ErrInvalid)
	}
	if req.url == "" {
		return fmt.Errorf("%w: url must not be empty", ErrInvalid)
	}
	url := req.url
	if len(req.query) > 0 {
		url += "?" + req.query.Encode()
	}
	hreq, err := http.NewRequestWithContext(ctx, req.method, url, req.body)
	if err != nil {
		return err
	}
	if req.contentType != "" {
		hreq.Header.Set("Content-Type", req.contentType)
	}
	for k, vals := range req.extraHeaders {
		for _, v := range vals {
			hreq.Header.Add(k, v)
		}
	}

	hres, err := client.Do(hreq)
	if err != nil {
		return err
	}
	defer hres.Body.Close()

	data, _ := httputil.DumpResponse(hres, true)

	log.Printf("Response(raw): '%s'\n", data)

	return json.NewDecoder(hres.Body).Decode(res)
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

type ClientOption utils.Option

type clientOptions struct {
	accessKey string
	secretKey string
}

func WithCredentials(accessKey, secretKey string) ClientOption {
	return utils.OptionFn(func(op any) {
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
		transport = utils.NewAuthorizedTransport(options.accessKey, options.secretKey)
	}
	return &Client{
		Client: &http.Client{
			Transport: transport,
		},
	}
}
