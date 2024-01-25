package titannetwork

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/qbox/net-deftones/util"
)

const _TITANNETWORK_ENDPOINT = "http://gateway.titannetwork.cn"

var (
	ErrInvalid     = errors.New("titannetwork: invalid argument")
	ErrProtocol    = errors.New("titannetwork: protocol or network/io error")
	ErrUnavailable = errors.New("titannetwork: service unavailable")
)

// Client represents a titannetwork client
type Client struct {
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

func (client *Client) send(ctx context.Context, req *request, res any) error {
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

	if hres.StatusCode >= http.StatusBadRequest {
		var cause error
		switch {
		case hres.StatusCode >= http.StatusInternalServerError:
			cause = ErrUnavailable
		case hres.StatusCode >= http.StatusBadRequest:
			cause = ErrInvalid
		default:
			cause = ErrProtocol
		}
		return fmt.Errorf("%w: %s (raw request: %+v)", cause, hres.Status, hreq.URL)
	}

	return nil
}

type LogUrlRequest struct {
	Domain    string
	Timestamp time.Time
	Token     string
}

type LogUrlResponseV2 struct {
	Result  string
	Message string
	Urls    []string
}

func (client *Client) BossFlowLogUrlV2(ctx context.Context, req *LogUrlRequest) (*LogUrlResponseV2, error) {
	var res LogUrlResponseV2
	var payload = make(url.Values)

	payload.Add("domain", req.Domain)
	payload.Add("time", req.Timestamp.Format("200601021504"))
	payload.Add("token", req.Token)

	if err := util.WithRetry(ctx, func() error {
		err := client.send(ctx, &request{
			path:   "/boss/flow/log_url/v2",
			method: http.MethodPost,
			form:   payload,
		}, &res)
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalid):
				return errors.Join(err, context.Canceled)
			default:
				return err
			}
		}
		return nil
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
	return &res, nil
}

type LogUrlResponseV1 struct {
	Result  string
	Message string
	Url     string
}

func (client *Client) BossFlowLogUrlV1(ctx context.Context, req *LogUrlRequest) (*LogUrlResponseV1, error) {
	var res LogUrlResponseV1
	var payload = make(url.Values)

	payload.Add("domain", req.Domain)
	payload.Add("time", req.Timestamp.Format("200601021504"))

	if err := util.WithRetry(ctx, func() error {
		err := client.send(ctx, &request{
			path:   "/boss/flow/log_url/v1",
			method: http.MethodPost,
			form:   payload,
		}, &res)
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalid):
				return errors.Join(err, context.Canceled)
			default:
				return err
			}
		}
		return nil
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
	return &res, nil
}

func (client *Client) BossFlowDomainList(ctx context.Context, since time.Time) ([]string, error) {
	var res struct {
		Result  string
		Message string
		Domains []string
	}

	if err := client.send(ctx, &request{path: "/boss/flow/domain_list/v1"}, &res); err != nil {
		return nil, err
	}
	return res.Domains, nil
}

type Option util.Option

type clientOptions struct {
	endpoint string
	username string
	secret   []byte
}

func WithCredential(username string, secret []byte) Option {
	return util.OptionFunc(func(opt any) {
		opt.(*clientOptions).username = username
		opt.(*clientOptions).secret = secret
	})
}

func WithEndpoint(endpoint string) Option {
	return util.OptionFunc(func(opt any) { opt.(*clientOptions).endpoint = endpoint })
}

func NewClient(opts ...Option) *Client {
	var options clientOptions

	for _, op := range opts {
		op.Apply(&options)
	}
	if options.endpoint == "" {
		options.endpoint = _TITANNETWORK_ENDPOINT
	}
	return &Client{
		Client: &http.Client{Transport: &taiwuTransport{
			username: options.username,
			secret:   options.secret,
		}},
		endpoint: options.endpoint,
	}
}
