package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"

	"github.com/qbox/net-deftones/logger"
)

type HttpClient interface {
	Get(ctx context.Context, path string, res any, options ...Option) error
	Post(ctx context.Context, path string, res any, options ...Option) error
}

type httpOptions struct {
	body          []byte
	contentType   string
	params        map[string]any
	authorization string
}
type Option func(*httpOptions)

func WithBody(body []byte, contentType string) Option {
	return func(opts *httpOptions) {
		opts.contentType = contentType
		if strings.Contains(contentType, "application/json") {
			var err error
			opts.body, err = json.Marshal(body)
			if err != nil {
				panic("json.Marshal error: " + err.Error())
			}
		} else {
			opts.body = body
		}
		if contentType == "" {
			opts.contentType = "text/plain; charset=utf-8"
		}
	}
}

var ErrInvalid = errors.New("http: Invalid argument")

type client struct {
	*http.Client
	host string
}

func newClient(host string) *client {
	return &client{host: host, Client: http.DefaultClient}
}

type Response = any

// Get send a http GET path request. resp should be a pointer to the response
func (cli *client) Get(ctx context.Context, path string, resp Response, options ...Option) error {
	req, err := cli.newRequest(ctx, http.MethodGet, path, options...)
	if err != nil {
		return fmt.Errorf("client: newRequest: %v", err)
	}
	return cli.send(ctx, req, resp)
}

func (cli *client) Post(ctx context.Context, path string, res Response, options ...Option) error {
	req, err := cli.newRequest(ctx, http.MethodPost, path, options...)
	if err != nil {
		return fmt.Errorf("client.newRequest: %v", err)
	}
	return cli.send(ctx, req, res)
}

func (cli *client) newRequest(ctx context.Context, method, path string, options ...Option) (*http.Request, error) {
	var opts httpOptions
	var query url.Values
	var body io.Reader

	if cli.host == "" {
		return nil, ErrInvalid
	}
	for _, op := range options {
		op(&opts)
	}
	if len(opts.params) > 0 {
		query = make(url.Values)
		for key, value := range opts.params {
			query.Set(key, fmt.Sprint(value))
		}
	}
	host := cli.host
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + cli.host
	}
	URL := host + path
	if len(query) > 0 {
		URL += "?" + query.Encode()
	}
	if len(opts.body) > 0 {
		if opts.contentType == "" {
			return nil, fmt.Errorf("Content-Type is empty but body not empty")
		}
		body = bytes.NewReader(opts.body)
	}
	req, err := http.NewRequestWithContext(ctx, method, URL, body)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %v", err)
	}
	if len(opts.body) > 0 && opts.contentType != "" {
		req.Header.Set("Content-Type", opts.contentType)
	}
	if opts.authorization != "" {
		req.Header.Set("Authorization", opts.authorization)
	}
	return req, nil
}

const RESPONSE_BODY_MAX = 1 << 20

func (cli *client) send(ctx context.Context, req *http.Request, resp any) error {
	res, err := cli.Do(req)
	if err != nil {
		return fmt.Errorf("http.Client.Do: %v", err)
	}
	defer res.Body.Close()

	if length := res.ContentLength; length != -1 && length > RESPONSE_BODY_MAX {
		return fmt.Errorf("The Content-Length %d is greater than RESPONSE_BODY_MAX %d",
			res.ContentLength, RESPONSE_BODY_MAX)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("io.ReadAll(body): %v", err)
	}
	switch contentType := res.Header.Get("Content-Type"); {
	case strings.Contains(contentType, "application/json"):
		return json.Unmarshal(data, resp)

	case strings.Contains(contentType, "application/xml"):
		return xml.Unmarshal(data, resp)

	case strings.Contains(contentType, "text/"):
		_type := reflect.TypeOf(resp)
		if _type.Kind() == reflect.Pointer && _type.Elem().Kind() == reflect.String {
			reflect.ValueOf(resp).Elem().Set(reflect.ValueOf(string(data)))
		} else {
			return fmt.Errorf("expected a pointer type of `resp` param, but: %v", _type.Elem().Kind())
		}
	default:
		return fmt.Errorf("can't read response body: `%v`", data)
	}
	return nil
}

func main() {
	var result string
	ctx := logger.NewContext(context.Background(), logger.New())
	err := newClient("https://www.baidu.com").Get(ctx, "/", &result)
	if err != nil {
		fatal("client.Get error: %v\n", err)
	}

	fmt.Printf("result: %v\n", result[:100])
}

func fatal(format string, v ...any) {
	fmt.Fprintf(os.Stderr, format, v...)
	os.Exit(1)
}
