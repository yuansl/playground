package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/yuansl/playground/util"
)

var fatal = util.Fatal

type hookTransport struct{}

// RoundTrip implements http.RoundTripper.
func (t *hookTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	data, err := httputil.DumpRequest(req, true)
	fmt.Printf("Request(raw): '%s'\n", data)

	res, err := http.DefaultTransport.RoundTrip(req)

	data, err = httputil.DumpResponse(res, false)
	fmt.Printf("Response(raw): '%s'\n", data)

	return res, err
}

var _ http.RoundTripper = (*hookTransport)(nil)

func NewTransport() http.RoundTripper {
	return &hookTransport{}
}

type metadata map[string]any

type Option util.Option

type options struct {
	headers http.Header
}

func WithHeader(header http.Header) Option {
	return util.OptionFunc(func(opt any) {
		opt.(*options).headers = header
	})
}

func send(ctx context.Context, client *http.Client, URL, method string, opts ...Option) metadata {
	var options options

	for _, opt := range opts {
		opt.Apply(&options)
	}
	req, err := http.NewRequest(method, URL, nil)
	if err != nil {
		fatal("http.NewRequest:", err)
	}
	for k, v := range options.headers {
		for _, it := range v {
			req.Header.Add(k, it)
		}
	}
	res, err := client.Do(req)
	if err != nil {
		fatal("http.DefaultClient.Do:", err)
	}
	defer res.Body.Close()

	io.Copy(io.Discard, res.Body)

	var meta = make(metadata)

	for k, v := range res.Header {
		if len(v) > 0 {
			meta[k] = v[0]
		}
	}
	return meta
}

func main() {
	URL := "http://ria8j59xt.hd-bkt.clouddn.com/dnlivestream/2023-12-03-01/qn-pcdngw.cdn.huya.com/qn-pcdngw.cdn.huya.com_0_part-00043-e390249e-0d4a-46c2-a914-dabc77c283d7.c000.json.gz.gz?e=1707635374&token=V5BwWT7pVm1S_EVHt2bfg4qOS-1VDLXCo1k6MqN1:-Q5PyFoj4xBxcA3wKYcZBRGaWVU="
	client := &http.Client{Transport: NewTransport()}

	ctx := context.TODO()

	meta := send(ctx, client, URL, http.MethodHead)

	if rangebytes := meta["Accept-Ranges"]; strings.TrimSpace(rangebytes.(string)) == "bytes" {
		meta = send(ctx, client, URL, http.MethodGet, WithHeader(http.Header{"Range": []string{"bytes=2605711360-"}}))
		fmt.Printf("Response=%v\n", meta)
	}
}
