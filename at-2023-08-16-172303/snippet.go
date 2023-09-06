// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-08-16 17:23:03

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"runtime"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
)

const _ENDPOINT_DEFAULT = "http://cs50:20052"

func fatal(v ...any) {
	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	fmt.Fprintf(os.Stderr, "%s:\n\t%s:%d\n", fn.Name(), file, line)
	os.Exit(1)
}

type Client struct {
	*http.Client
	host string
	cred *auth.Credentials
}

func NewClient(cred *auth.Credentials) *Client {
	return &Client{
		host: _ENDPOINT_DEFAULT,
		Client: &http.Client{
			Transport: &TransportWithAuthorization{
				r:    http.DefaultTransport,
				cred: cred,
			}},
	}
}

type TransportWithAuthorization struct {
	r    http.RoundTripper
	cred *auth.Credentials
}

func (authT *TransportWithAuthorization) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := authT.cred.AddToken(auth.TokenQiniu, req); err != nil {
		return nil, err
	}
	fmt.Printf("Authorization='%s'\n", req.Header.Get("Authorization"))
	return authT.r.RoundTrip(req)
}

type Request struct {
	Body io.Reader
}

type HttpRequest struct {
	*Request
	Path        string
	Method      string
	Params      url.Values
	Body        io.Reader
	ContentType string
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (*Client) GetCdnTrafficStat(ctx context.Context) {}

func run(client *Client, payload any) {
	body := bytes.NewBuffer(nil)

	json.NewEncoder(body).Encode(payload)

	req, err := http.NewRequest(http.MethodPost, client.host+"/v2/cdn/traffic/stat", body)
	if err != nil {
		fatal("http.NewRequest:", err)
	}
	req.Header.Set("Content-Type", "application/json")

	data, err := httputil.DumpRequest(req, true)
	fmt.Printf("Request(raw): '%s'\n", data)

	res, err := client.Do(req)
	if err != nil {
		fatal("http.Client.Do:", err)
	}
	defer res.Body.Close()

	data, err = httputil.DumpResponse(res, res.StatusCode >= http.StatusBadRequest)
	if err != nil {
		fatal("httputil.DumpResonse:", err)
	}

	fmt.Printf("Response(raw): '%s'\n", data)

	var r struct {
		Result any `json:"result"`
		*Error `json:"error,omitempty"`
	}

	if err = json.NewDecoder(io.LimitReader(res.Body, 10<<20)).Decode(&r); err != nil {
		fatal("json.Decode error:", err)
	}

	fmt.Printf("Results: %+v\n", r)
}

func main() {
	credential := qbox.NewMac(os.Getenv("ACCESS_KEY"), os.Getenv("SECRET_KEY"))
	client := NewClient(credential)

	payload := map[string]any{"start": "2023-07-01T00:00:00Z", "end": "2023-07-30T00:00:00+00:00", "g": "day"}

	run(client, payload)
}
