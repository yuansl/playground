package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/qbox/net-deftones/logger"
	"github.com/yuansl/playground/util"
)

const _ETL_ENDPOINT_DEFAULT = "http://xs211:12324"

var ErrInvalid = errors.New("Invalid argument")

type httpRequest struct {
	uri         string
	method      string
	query       url.Values
	body        io.Reader
	contentType string
}

func sendRequest(ctx context.Context, req *httpRequest, res any) error {
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

	hres, err := http.DefaultClient.Do(hreq)
	if err != nil {
		return err
	}
	defer hres.Body.Close()

	data, _ := httputil.DumpResponse(hres, hres.StatusCode >= http.StatusBadRequest)

	logger.FromContext(ctx).Warnf("Response(raw): '%s'\n", data)

	return json.NewDecoder(hres.Body).Decode(res)
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

type EtlRetryResponse []string

func SendEtlRetryRequest(ctx context.Context, req *EtlRetryRequest) (EtlRetryResponse, error) {
	var buf bytes.Buffer
	var hres EtlRetryResponse

	_ = json.NewEncoder(&buf).Encode(req)

	if err := sendRequest(ctx,
		&httpRequest{
			uri:         _ETL_ENDPOINT_DEFAULT + "/v5/etl/retry",
			method:      http.MethodPost,
			body:        &buf,
			contentType: "application/json",
		}, &hres); err != nil {
		return nil, fmt.Errorf("sendRequest: %v", err)
	}
	return hres, nil
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

func SendDaySyncsRequest(ctx context.Context, req *DaySyncsRequest) (DaySyncsResponse, error) {
	var hres any
	var query = make(url.Values)

	query.Set("domains", strings.Join(req.Domains, ","))
	query.Set("start", req.Start.Format(time.DateOnly))
	query.Set("end", req.End.Format(time.DateOnly))
	query.Set("type", strconv.Itoa(int(req.Type)))
	query.Set("cdn", "all")
	query.Set("force", "true")
	query.Set("sink", "true")

	if err := sendRequest(ctx,
		&httpRequest{
			uri:    "http://jjh2746:12323/v3/unify/day/syncs",
			method: http.MethodGet,
			query:  query,
		}, &hres); err != nil {
		return nil, fmt.Errorf("SendDaySyncsRequest: sendRequest: %v => '%v'", err, hres)
	}
	return *(*DaySyncsResponse)(&hres), nil
}

const _SYNC_BATCHSIZE_MAX = 200

func Sync(ctx context.Context, domains []string, start, end time.Time) error {
	for i := 0; i < _SYNC_BATCHSIZE_MAX; i += _SYNC_BATCHSIZE_MAX {
		i0, i1 := i, i+_SYNC_BATCHSIZE_MAX
		if i1 > len(domains) {
			i1 = len(domains)
		}
		res, err := SendDaySyncsRequest(ctx, &DaySyncsRequest{
			Domains: domains[i0:i1],
			Start:   start,
			End:     end,
		})
		if err != nil {
			return fmt.Errorf("SendDaySyncsRequest: %v => '%v'", err, res)
		}
	}
	return nil
}

const ETL_RETRY_INTERVAL_MAX = 7 * 24 * time.Hour

func EtlRetry(ctx context.Context, domains []string, start, end time.Time, cdn string) error {
	if start.IsZero() || end.IsZero() {
		return fmt.Errorf("%w: start or end must not be zero", ErrInvalid)
	}

	for start0 := start; start0.Before(end); start0 = start0.Add(ETL_RETRY_INTERVAL_MAX) {
		end0 := start0.Add(ETL_RETRY_INTERVAL_MAX)
		if end0.After(end) {
			end0 = end
		}

		res, err := SendEtlRetryRequest(ctx, &EtlRetryRequest{
			Cdn:          cdn,
			Domains:      domains,
			Start:        start0,
			End:          end0,
			Force:        true,
			Manual:       true,
			IgnoreOnline: true,
		})
		if err != nil {
			return fmt.Errorf("EtlRetryRequest: %v => '%+v'", err, res)
		}
		logger.FromContext(ctx).Infof("domains=%s,start=%s,end=%s,cdn=%s,etl result: %+v\n", domains, start0, end0, cdn, res)
	}

	return nil
}

var _options struct {
	filename string
	domains  string
	mode     string
	start    time.Time
	end      time.Time
	cdn      string
}

func parseCmdArgs() {
	flag.StringVar(&_options.filename, "f", "", "specify domains file (csv)")
	flag.StringVar(&_options.domains, "domains", "", "specify domains(seperate by comma)")
	flag.StringVar(&_options.mode, "mode", "etl", "specify operation type. etl|sync")
	flag.TextVar(&_options.start, "start", time.Time{}, "specify start time in format RFC3339")
	flag.TextVar(&_options.end, "end", time.Time{}, "specify end time in format RFC3339")
	flag.StringVar(&_options.cdn, "cdn", "", "specify cdn provider")
	flag.Parse()

	if _options.domains == "" {
		fmt.Fprintf(os.Stderr, "domains must not be empty\n")
		os.Exit(1)
	}

	if _options.cdn == "" {
		fmt.Fprintf(os.Stderr, "cdn must not be empty\n")
		os.Exit(1)
	}
}

var fatal = util.Fatal

func loadDomains(filename string) []string {
	var domains []string

	fp, err := os.Open(filename)
	if err != nil {
		fatal("os.Open:", err)
	}
	for r := bufio.NewReader(fp); ; {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fatal("bufio.ReadBytes:", err)
			}
			break
		}
		line = bytes.TrimSpace(line)

		if bytes.Equal(line, []byte("domain")) {
			continue
		}
		domains = append(domains, string(line))
	}
	return domains
}

func main() {
	parseCmdArgs()

	domains := strings.Split(_options.domains, ",")
	ctx := logger.NewContext(context.Background(), logger.New())

	switch _options.mode {
	case "etl":
		if err := EtlRetry(ctx, domains, _options.start, _options.end, _options.cdn); err != nil {
			fatal(err)
		}
	case "sync":
		Sync(ctx, domains, _options.start, _options.end)
	default:
		fatal("Unknown mode: " + _options.mode)
	}
}
