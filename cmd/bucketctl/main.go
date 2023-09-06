package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/qiniu/go-sdk/v7/auth"
)

const _FUSION_SUFY_ENDPOINT_DEFAULT = "http://fusiondomainproxy.qcdn-sophon-sufy.qa.qiniu.io"

type AuthorizedTransport struct {
	http.RoundTripper
	creds *auth.Credentials
}

// RoundTrip implements http.RoundTripper.
func (t *AuthorizedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.creds.AddToken(auth.TokenQiniu, req)

	data, _ := httputil.DumpRequest(req, true)
	log.Printf("Request(raw): '%s'\n", data)
	return t.RoundTripper.RoundTrip(req)
}

var _ http.RoundTripper = (*AuthorizedTransport)(nil)

type Client struct {
	*http.Client
}

func NewClient(accessKey, secretKey string) *Client {
	return &Client{Client: &http.Client{
		Transport: &AuthorizedTransport{
			RoundTripper: http.DefaultTransport,
			creds:        auth.New(accessKey, secretKey),
		},
	}}
}

type httpRequest struct {
	uri          string
	method       string
	query        url.Values
	body         io.Reader
	contentType  string
	extraHeaders http.Header
}

func (client *Client) sendRequest(ctx context.Context, req *httpRequest, res any) error {
	url := req.uri
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

func fatal(v ...any) {
	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	fmt.Fprintf(os.Stderr, "%s:\n\t%s:%d\n", fn.Name(), file, line)
	os.Exit(1)
}

func GetBucket(ctx context.Context, domain string, time time.Time, client *Client) string {
	var res struct {
		Bucket string
	}
	var extraHeader = make(http.Header)
	extraHeader.Set("Account", "defy@qiniu.com")
	var query = make(url.Values)

	query.Add("domain", domain)
	if !time.IsZero() {
		query.Add("time", strconv.Itoa(int(time.Unix())))
	}
	if err := client.sendRequest(ctx,
		&httpRequest{
			uri:          _FUSION_SUFY_ENDPOINT_DEFAULT + "/admin/domain/deliverybucket",
			query:        query,
			extraHeaders: extraHeader,
		}, &res); err != nil {
		fatal("sendRequest:", err)
	}

	return res.Bucket
}

var globalOptions struct {
	accessKey string
	secretKey string
	domain    string
	time      time.Time
}

func parseCmdArgs() {
	flag.StringVar(&globalOptions.accessKey, "ak", os.Getenv("ACCESS_KEY"), "qiniu account access key")
	flag.StringVar(&globalOptions.secretKey, "sk", os.Getenv("SECRET_KEY"), "qiniu account secret key")
	flag.StringVar(&globalOptions.domain, "domain", "www.example.com", "specify a domain")
	flag.TextVar(&globalOptions.time, "time", time.Time{}, "specify a effective time")
	flag.Parse()
}

func main() {
	parseCmdArgs()

	bucket := GetBucket(context.TODO(), globalOptions.domain, globalOptions.time, NewClient(globalOptions.accessKey, globalOptions.secretKey))
	fmt.Printf("bucket of domain %s: %q\n", globalOptions.domain, bucket)
}
