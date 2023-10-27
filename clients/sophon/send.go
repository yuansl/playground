package sophon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type httpRequest struct {
	url          string
	method       string
	query        url.Values
	body         io.Reader
	contentType  string
	extraHeaders http.Header
}

func (client *Client) send(ctx context.Context, req *httpRequest, res any) error {
	if req.method == "" {
		return fmt.Errorf("%w: method must not be empty", ErrInvalid)
	}
	if req.url == "" {
		return fmt.Errorf("%w: url must not be empty", ErrInvalid)
	}
	url := req.url
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
