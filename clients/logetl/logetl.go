package logetl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	_LOGETL_ENDPOINT       = "http://xs211:12324"
	_TRAFFIC_SINK_ENDPOINT = "http://jjh2746:12323"
)

var ErrInvalid = errors.New("clients.logetl: Invalid argument")

type Client struct {
	*http.Client
}

type httpRequest struct {
	url         string
	method      string
	query       url.Values
	body        io.Reader
	contentType string
}

func (client *Client) sendRequest(ctx context.Context, req *httpRequest, res any) error {
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

	data, _ := httputil.DumpRequest(hreq, false)
	log.Printf("Request(raw): '%s'\n", data)

	hres, err := client.Do(hreq)
	if err != nil {
		return err
	}
	defer hres.Body.Close()

	data, _ = httputil.DumpResponse(hres, hres.StatusCode >= http.StatusBadRequest)

	log.Printf("Response(raw): '%s'\n", data)

	if strings.Contains(hres.Header.Get("Content-Type"), "application/json") {
		return json.NewDecoder(hres.Body).Decode(res)
	}
	_, err = io.Copy(io.Discard, hres.Body)
	return err
}

type EtlTaskRequest struct {
	Id string
}

type EtlTaskResponse struct {
	Result struct {
		Id          string   `json:"id"`
		Domain      string   `json:"domain"`
		Hour        string   `json:"hour"`
		Status      string   `json:"status"`
		FinalStatus string   `json:"finalStatus"`
		Messages    []string `json:"messages"`
		Cdn         string   `json:"cdn"`
		OriginCdn   string   `json:"originCdn"`
		MessageId   string   `json:"messageId"`
	} `json:"rawTask"`
}

func (*Client) GetEtlTasks(r *EtlTaskRequest) (*EtlTaskResponse, error) {
	var payload EtlTaskResponse

	res, err := http.Get("http://xs201:12324/v5/etl/tasks/" + r.Id)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&payload)

	return &payload, err
}

type EtlTasksRequest struct {
	Cdn           string    `json:"cdn"`
	Domain        string    `json:"domain"`
	Hour          time.Time `json:"-"`
	Manual        bool      `json:"manual"`
	Force         bool      `json:"force"`
	Priority      int       `json:"priority"`
	NotifyUpload  bool      `json:"notifyUpload"`
	NotifyTraffic bool      `json:"notifyTraffic"`
	NotifyAnalyze bool      `json:"notifyAnalyze"`
	IgnoreOnline  bool      `json:"ignoreOnline"`
}

func (r *EtlTasksRequest) MarshalJSON() ([]byte, error) {
	type alias EtlTasksRequest
	x := struct {
		*alias
		Hour string `json:"hour"`
	}{
		alias: (*alias)(r),
		Hour:  r.Hour.Format("2006-01-02-15"),
	}
	return json.Marshal(x)
}

type EtlTasksResponse struct {
	TaskId string
}

func (client *Client) SendEtlTasksRequest(ctx context.Context, req *EtlTasksRequest) (*EtlTasksResponse, error) {
	var res any
	var buf bytes.Buffer

	_ = json.NewEncoder(&buf).Encode(req)

	if err := client.sendRequest(ctx, &httpRequest{
		url:         _LOGETL_ENDPOINT + "/v5/etl/tasks",
		method:      http.MethodPost,
		contentType: "application/json",
		body:        &buf,
	}, &res); err != nil {
		return nil, err
	}
	return &EtlTasksResponse{TaskId: res.(string)}, nil
}

type EtlRetryRequest struct {
	Cdn          string    `json:"cdn"`
	Domains      []string  `json:"domains"`
	Force        bool      `json:"force"`
	Manual       bool      `json:"manual"`
	Start        time.Time `json:"start"`
	End          time.Time `json:"end"`
	IgnoreOnline bool      `json:"ignoreOnline"`
}

func (r *EtlRetryRequest) MarshalJSON() ([]byte, error) {
	type Alias EtlRetryRequest

	var x = struct {
		*Alias
		Start string `json:"start"`
		End   string `json:"end"`
	}{
		Alias: (*Alias)(r),
		Start: r.Start.Format("2006-01-02-15"),
		End:   r.End.Format("2006-01-02-15"),
	}
	return json.Marshal(x)
}

type EtlRetryResponse any

func (client *Client) SendEtlRetryRequest(ctx context.Context, req *EtlRetryRequest) (*EtlRetryResponse, error) {
	var buf bytes.Buffer
	var hres any

	_ = json.NewEncoder(&buf).Encode(req)

	if err := client.sendRequest(ctx,
		&httpRequest{
			url:         _LOGETL_ENDPOINT + "/v5/etl/retry",
			method:      http.MethodPost,
			body:        &buf,
			contentType: "application/json",
		}, &hres); err != nil {
		return nil, fmt.Errorf("sendRequest: %v", err)
	}
	res := hres.(EtlRetryResponse)
	return &res, nil
}

type DataType int

const (
	DataTypeBandwidth DataType = iota // 0: bandwidth
)

type DaySyncsRequest struct {
	Domains []string
	Start   time.Time
	End     time.Time
	Type    DataType
	Cdn     string
	Force   bool
	Sink    bool
}

type DaySyncsResponse any

func (client *Client) GetUnifyDaySyncs(ctx context.Context, req *DaySyncsRequest) (DaySyncsResponse, error) {
	var hres any
	var query = make(url.Values)

	query.Set("domains", strings.Join(req.Domains, ","))
	query.Set("start", req.Start.Format(time.DateOnly))
	query.Set("end", req.End.Format(time.DateOnly))
	query.Set("type", strconv.Itoa(int(req.Type)))
	if req.Cdn == "" {
		req.Cdn = "all"
	}
	query.Set("cdn", req.Cdn)
	query.Set("force", "true")
	query.Set("sink", "true")

	if err := client.sendRequest(ctx,
		&httpRequest{
			url:    _TRAFFIC_SINK_ENDPOINT + "/v3/unify/day/syncs",
			method: http.MethodGet,
			query:  query,
		}, &hres); err != nil {
		return nil, fmt.Errorf("SendDaySyncsRequest: sendRequest: %v => '%v'", err, hres)
	}
	return *(*DaySyncsResponse)(&hres), nil
}

func NewClient() *Client {
	return &Client{Client: http.DefaultClient}
}
