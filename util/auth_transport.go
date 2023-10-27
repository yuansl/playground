package util

import (
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/qiniu/go-sdk/v7/auth"
)

type AuthorizedTransport struct {
	http.RoundTripper
	creds     *auth.Credentials
	tokenType auth.TokenType
}

// RoundTrip implements http.RoundTripper.
func (t *AuthorizedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.creds.AddToken(t.tokenType, req)

	data, _ := httputil.DumpRequest(req, true)
	log.Printf("Request(raw): %q\n", data)

	return t.RoundTripper.RoundTrip(req)
}

var _ http.RoundTripper = (*AuthorizedTransport)(nil)

type TransportOption Option

type transportOptions struct {
	tokenType auth.TokenType
	transport http.RoundTripper
}

func WithTokenType(tokenType auth.TokenType) TransportOption {
	return OptionFn(func(op any) {
		op.(*transportOptions).tokenType = tokenType
	})
}

func WithTransport(transport http.RoundTripper) TransportOption {
	return OptionFn(func(op any) {
		op.(*transportOptions).transport = transport
	})
}

func NewAuthorizedTransport(accessKey, secretKey string, opts ...TransportOption) http.RoundTripper {
	var options transportOptions

	for _, op := range opts {
		op.Apply(&options)
	}
	if options.tokenType < 0 || options.tokenType > auth.TokenQBox {
		options.tokenType = auth.TokenQiniu
	}
	if options.transport == nil {
		options.transport = http.DefaultTransport
	}
	return &AuthorizedTransport{
		RoundTripper: options.transport,
		creds:        auth.New(accessKey, secretKey),
		tokenType:    options.tokenType,
	}
}
