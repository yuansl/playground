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
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
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

	log.Printf("Response(raw): '%s'\n", data)

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

type EtlRetryResponse any

func SendEtlRetryRequest(ctx context.Context, req *EtlRetryRequest) (*EtlRetryResponse, error) {
	var buf bytes.Buffer
	var hres any

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

func EtlRetry(ctx context.Context, domains []string, start, end time.Time) error {
	if start.IsZero() || end.IsZero() {
		return fmt.Errorf("%w: start or end must not be zero", ErrInvalid)
	}

	if res, err := SendEtlRetryRequest(ctx, &EtlRetryRequest{
		Cdn:          "qiniucdn",
		Domains:      domains,
		Start:        start,
		End:          end,
		Force:        true,
		Manual:       true,
		IgnoreOnline: true,
	}); err != nil {
		return fmt.Errorf("EtlRetryRequest: %v => '%+v'", err, res)
	}
	return nil
}

var (
	filename string
	mode     string
	start    time.Time
	end      time.Time
)

func parseCmdArgs() {
	flag.StringVar(&filename, "f", "", "specify domains file (csv)")
	flag.StringVar(&mode, "mode", "etl", "specify operation type. etl|sync")
	flag.TextVar(&start, "start", time.Time{}, "specify start time in format RFC3339")
	flag.TextVar(&end, "end", time.Time{}, "specify end time in format RFC3339")
	flag.Parse()
}

func fatal(v ...any) {
	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	fmt.Fprintf(os.Stderr, "%s:\n\t%s:%d\n", fn.Name(), file, line)
	os.Exit(1)
}

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

	domains := loadDomains(filename)

	switch mode {
	case "etl":
		if err := EtlRetry(context.TODO(), domains, start, end); err != nil {
			fatal(err)
		}

	case "sync":
		Sync(context.TODO(), domains, start, end)
	default:
		fatal("Unknown mode: " + mode)
	}
}
