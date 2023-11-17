package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/qbox/net-deftones/util"
	"github.com/yuansl/playground/logger"
)

var (
	ErrInvalid     = errors.New("taiwulogctl: invalid argument")
	ErrProtocol    = errors.New("taiwulogctl: protocol or network/io error")
	ErrUnavailable = errors.New("taiwulogctl: service unavailable")
)

type Link struct {
	Url string
}

type TaiwuService interface {
	LogLink(ctx context.Context, domain string, timestamp time.Time) ([]Link, error)
}

type taiwuTransport struct {
	username string
	secret   string
}

// RoundTrip implements http.RoundTripper.
func (taiwu *taiwuTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	form, err := url.ParseQuery(string(data))
	if err != nil {
		return nil, err
	}
	form.Add("username", taiwu.username)
	form.Add("secret", taiwu.secret)
	content := []byte(form.Encode())
	req.Body = io.NopCloser(bytes.NewReader(content))
	req.ContentLength = int64(len(content))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	data, _ = httputil.DumpRequest(req, true)
	logger.New().Infof("http request(raw): '%s'\n", bytes.Replace(data, []byte("\r\n"), []byte("..."), -1))

	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	data, _ = httputil.DumpResponse(res, res.StatusCode >= http.StatusInternalServerError)
	logger.New().Infof("http response(raw): '%s'\n", bytes.Replace(data, []byte("\r\n"), []byte("..."), -1))

	return res, nil
}

var _ http.RoundTripper = (*taiwuTransport)(nil)

type taiwuClient struct {
	*http.Client
	endpoint string
	token    string
}

type request struct {
	path   string
	method string // GET, POST ...
	query  url.Values
	form   url.Values
	body   io.Reader
}

func (client *taiwuClient) send(ctx context.Context, req *request, res any) error {
	URL := client.endpoint + req.path
	if len(req.query) > 0 {
		URL += "?" + req.query.Encode()
	}
	body := req.body
	if body == nil && len(req.form) > 0 {
		body = bytes.NewBufferString(req.form.Encode())
	}

	hreq, err := http.NewRequestWithContext(ctx, req.method, URL, body)
	if err != nil {
		return fmt.Errorf("%w: http.NewRequest: %v", ErrInvalid, err)
	}
	hres, err := client.Do(hreq)
	if err != nil {
		return fmt.Errorf("%w: http.Client.Do: %v", ErrProtocol, err)
	}
	defer hres.Body.Close()

	if contentType := hres.Header.Get("Content-Type"); strings.Contains(contentType, "application/json") {
		if err = json.NewDecoder(hres.Body).Decode(res); err != nil {
			return fmt.Errorf("%w: json.Decode: %v", ErrProtocol, err)
		}
	}
	switch {
	case hres.StatusCode >= http.StatusInternalServerError:
		return fmt.Errorf("%w: %s", ErrUnavailable, hres.Status)
	case hres.StatusCode >= http.StatusBadRequest:
		return fmt.Errorf("%w: %s", ErrInvalid, hres.Status)
	default:
	}
	return nil
}

type LogUrlResponseV2 struct {
	Result  string
	Message string
	Urls    []string
}

func (client *taiwuClient) LogUrlV2(ctx context.Context, domain string, timestamp time.Time) (links []Link, err error) {
	payload := url.Values{}
	payload.Add("domain", domain)
	payload.Add("time", timestamp.Format("200601021504"))
	payload.Add("token", client.token)

	var res LogUrlResponseV2

	if err := util.WithRetry(ctx, func() error {
		return client.send(ctx, &request{
			path:   "/boss/flow/log_url/v2",
			method: http.MethodPost,
			form:   payload,
		}, &res)
	}); err != nil {
		return nil, err
	}
	if len(res.Urls) == 0 {
		switch res.Result {
		case "LogNotFound":
			return nil, fmt.Errorf("%w: LogNotFound %s", ErrInvalid, res.Message)
		default:
			return nil, fmt.Errorf("%w: %s %s", ErrProtocol, res.Result, res.Message)
		}
	}

	for _, it := range res.Urls {
		links = append(links, Link{Url: it})
	}
	return links, nil
}

type LogUrlResponseV1 struct {
	Result  string
	Message string
	Url     string
}

func (client *taiwuClient) LogUrlV1(ctx context.Context, domain string, timestamp time.Time) ([]Link, error) {
	payload := url.Values{}
	payload.Add("domain", domain)
	payload.Add("time", timestamp.Format("200601021504"))

	var res LogUrlResponseV1

	if err := util.WithRetry(ctx, func() error {
		return client.send(ctx, &request{
			path:   "/boss/flow/log_url/v1",
			method: http.MethodPost,
			form:   payload,
		}, &res)
	}); err != nil {
		return nil, err
	}
	if res.Url == "" {
		switch res.Result {
		case "LogNotFound":
			return nil, fmt.Errorf("%w: taiwu: LogNotFound %s", ErrInvalid, res.Message)
		default:
			return nil, fmt.Errorf("%w: %s %s", ErrProtocol, res.Result, res.Message)
		}
	}
	return []Link{{Url: res.Url}}, nil
}

// LogLink implements TaiwuService.
func (client *taiwuClient) LogLink(ctx context.Context, domain string, timestamp time.Time) ([]Link, error) {
	switch _version {
	case 1:
		return client.LogUrlV1(ctx, domain, timestamp)
	case 2:
		fallthrough
	default:
		return client.LogUrlV2(ctx, domain, timestamp)
	}
}

type DomainsResponse struct {
	Result  string
	Message string
	Domains []string
}

func (client *taiwuClient) Domains(ctx context.Context, since time.Time) ([]string, error) {
	var res DomainsResponse

	if err := client.send(ctx, &request{path: "/boss/flow/domain_list/v1"}, &res); err != nil {
		return nil, err
	}
	return res.Domains, nil
}

var _ TaiwuService = (*taiwuClient)(nil)

func NewTaiwuClient() TaiwuService {
	return &taiwuClient{
		Client: &http.Client{Transport: &taiwuTransport{
			username: "qiniu", secret: "a5c90e5370c80067a2ac78aab1badb90",
		}},
		endpoint: "http://gateway.titannetwork.cn",
		token:    "386BD183",
	}
}
