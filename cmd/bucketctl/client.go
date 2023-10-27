package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/yuansl/playground/util"
)

type Client struct {
	*http.Client
}

type Request struct {
	Uri          string
	Method       string
	Query        url.Values
	Body         io.Reader
	ContentType  string
	ExtraHeaders http.Header
}

func (client *Client) Send(ctx context.Context, req *Request, res any) error {
	url := req.Uri
	if len(req.Query) > 0 {
		url += "?" + req.Query.Encode()
	}
	hreq, err := http.NewRequestWithContext(ctx, req.Method, url, req.Body)
	if err != nil {
		return err
	}
	if hosts, exists := req.ExtraHeaders["Host"]; exists {
		hreq.Host = hosts[0]
	}
	if req.ContentType != "" {
		hreq.Header.Set("Content-Type", req.ContentType)
	}
	for k, vals := range req.ExtraHeaders {
		for _, v := range vals {
			hreq.Header.Set(k, v)
		}
	}
	data, _ := httputil.DumpRequest(hreq, true)
	fmt.Printf("Request(raw)='%s'\n", data)

	hres, err := client.Do(hreq)
	if err != nil {
		return err
	}
	defer hres.Body.Close()

	data, _ = httputil.DumpResponse(hres, true)
	log.Printf("Response(raw): '%s'\n", data)

	return json.NewDecoder(hres.Body).Decode(res)
}

func NewClient(accessKey, secretKey string) *Client {
	return &Client{
		Client: &http.Client{
			Transport: util.NewAuthorizedTransport(accessKey, secretKey),
		},
	}
}
