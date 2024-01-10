package titannetwork

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/qbox/net-deftones/logger"
)

type taiwuTransport struct {
	username string
	secret   []byte
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
	form.Add("secret", string(taiwu.secret))
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
	data, _ = httputil.DumpResponse(res, true) //FIXME: res.StatusCode >= http.StatusBadRequest)
	logger.New().Infof("http response(raw): '%s'\n", bytes.Replace(data, []byte("\r\n"), []byte("..."), -1))

	return res, nil
}

var _ http.RoundTripper = (*taiwuTransport)(nil)
