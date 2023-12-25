package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/yuansl/playground/util"
)

const (
	_LOGETL_ENDPOINT_DEFAULT = "http://xs211:12324"
	_LOGETL_TIMEFORMAT       = "2006-01-02-15"
)

type logetlClient struct {
	*http.Client
	endpoint string
}

// Retry implements LogetlService.
func (client *logetlClient) Retry(ctx context.Context, domains []string, cdn string, start time.Time, end time.Time) error {
	var body bytes.Buffer

	json.NewEncoder(&body).Encode(map[string]any{
		"cdn":          cdn,
		"domains":      domains,
		"start":        start.Local().Format(_LOGETL_TIMEFORMAT),
		"end":          end.Local().Format(_LOGETL_TIMEFORMAT),
		"manual":       true,
		"force":        true,
		"ignoreOnline": true,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, client.endpoint+"/v5/etl/retry", &body)
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http.Post: %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	fmt.Printf("etl/retry(domain='%v',hour='%s') result= '%s'\n", domains, start.Local().Format(time.DateTime), data)

	return nil
}

type BeforeRequestHook func(*http.Request)

type BeforeResponseHook func(*http.Request, *http.Response)

type hookedTransport struct {
	beforeSend     BeforeRequestHook
	beforeResponse BeforeResponseHook
}

func dumpHttpRequest(req *http.Request) {
	data, _ := httputil.DumpRequest(req, true)
	fmt.Printf("request(raw)='%s'\n", data)
}

func dumpHttpResponse(req *http.Request, res *http.Response) {
	data, _ := httputil.DumpResponse(res, res.StatusCode >= http.StatusBadRequest)
	fmt.Printf("response(raw)='%s'\n", data)
}

// RoundTrip implements http.RoundTripper.
func (t *hookedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.beforeSend != nil {
		t.beforeSend(req)
	}

	res, err := http.DefaultTransport.RoundTrip(req)

	if t.beforeResponse != nil {
		t.beforeResponse(req, res)
	}
	return res, err
}

var _ http.RoundTripper = (*hookedTransport)(nil)

func newTransport() http.RoundTripper {
	return &hookedTransport{
		beforeSend:     dumpHttpRequest,
		beforeResponse: dumpHttpResponse,
	}
}

var _ RetryUploader = (*logetlClient)(nil)

type LogetlOption util.Option

type logetlOptions struct{ endpoint string }

func WithLogetlEndpoint(endpoint string) LogetlOption {
	return util.OptionFunc(func(opt any) {
		opt.(*logetlOptions).endpoint = endpoint
	})
}

func NewLogetlService(opts ...LogetlOption) RetryUploader {
	var options logetlOptions
	for _, opt := range opts {
		opt.Apply(&options)
	}
	if options.endpoint == "" {
		options.endpoint = _LOGETL_ENDPOINT_DEFAULT
	}
	return &logetlClient{
		Client: &http.Client{
			Transport: newTransport(),
		},
		endpoint: options.endpoint,
	}
}
