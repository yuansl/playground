package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	gohttputil "net/http/httputil"

	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/util/httputil"
)

type Request struct {
	Path        string
	Method      string
	Body        []byte
	ContentType string
}

type Response = httputil.Response

type Client interface {
	Send(ctx context.Context, _ *Request) (*Response, error)
}

type httpClient struct {
	*http.Client
	endpoint string
}

// Do implements Client.
func (c *httpClient) Send(ctx context.Context, req *Request) (*Response, error) {
	var body bytes.Buffer
	if len(req.Body) > 0 {
		body.Write(req.Body)
	}
	URL := c.endpoint + req.Path
	hreq, err := http.NewRequestWithContext(ctx, req.Method, URL, &body)
	if err != nil {
		return nil, err
	}
	hreq.Header.Set("Content-Type", "application/json")

	data, _ := gohttputil.DumpRequest(hreq, true)
	logger.FromContext(ctx).Infof("Request(raw): `%s`\n", data)

	hres, err := c.Client.Do(hreq)
	if err != nil {
		return nil, err
	}
	var resp Response
	if err = json.NewDecoder(hres.Body).Decode(&resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

var _ Client = (*httpClient)(nil)
