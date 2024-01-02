package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"time"

	"github.com/qbox/net-deftones/util"
	"golang.org/x/sync/errgroup"
)

var (
	ErrInvalid  = errors.New("huyalog: invalid argument")
	ErrProtocol = errors.New("huyalog: network protocol/io error")
)

type uploaderService struct {
	uploaderCli *uploaderClient
	logQueryCli *logQueryClient
}

type metastat struct {
	Analyze int64 `json:"analyze"`
	Upload  int64 `json:"upload"`
	Store   int64 `json:"store"`
	Log     int64 `json:"log"`
	API     int64 `json:"api"`
}

func (srv *uploaderService) fetchOneHourUploadStat(ctx context.Context, domain, cdn string, hour time.Time) ([]UploadStat, error) {
	var stats []UploadStat
	result, err := srv.logQueryCli.LogFluxDomainCdnHour(ctx, &DomainCdnHourRequest{
		Domain: domain,
		Cdn:    cdn,
		Hour:   hour,
	})
	if err != nil {
		return nil, err
	}
	for m := 1; m <= 12; m++ {
		start := hour.Add(time.Duration(m-1) * 5 * time.Minute)
		end := start.Add(5 * time.Minute)
		stats = append(stats, UploadStat{
			Domain: domain,
			Cdn:    cdn,
			Bytes:  result.PerMinuteTotal[strconv.Itoa(m)].Upload,
			Start:  start,
			End:    end,
		})
	}
	return stats, nil
}

// Stat implements CdnlogUploader.
func (srv *uploaderService) Stat(ctx context.Context, domains []string, cdn string, begin time.Time, end time.Time) ([]UploadStat, error) {
	var stats []UploadStat
	var statsq = make(chan []UploadStat)

	go func() {
		defer close(statsq)

		wg, ctx := errgroup.WithContext(ctx)
		wg.SetLimit(runtime.NumCPU())

		for _, domain := range domains {
			_domain := domain
			for hour := begin; hour.Before(end); hour = hour.Add(1 * time.Hour) {
				_hour := hour
				wg.Go(func() error {
					var stats []UploadStat
					var err error

					util.WithRetry(ctx, func() error {
						stats, err = srv.fetchOneHourUploadStat(ctx, _domain, cdn, _hour)
						return err
					})
					if err != nil {
						return err
					}
					statsq <- stats
					return nil
				})
			}
		}
		wg.Wait()
	}()
	for stats0 := range statsq {
		stats = append(stats, stats0...)
	}
	return stats, nil
}

func (srv *uploaderService) Retry(ctx context.Context, domains []string, cdn string, start time.Time, end time.Time, manual bool) error {
	for _, domain := range domains {
		for hour := start; hour.Before(end); hour = hour.Add(1 * time.Hour) {
			if err := srv.uploaderCli.UploadRetryUpload(ctx, &UploadRequest{Cdn: cdn, Domain: domain, Time: hour}); err != nil {
				return err
			}
		}
	}
	return nil
}

type uploaderClient struct {
	*http.Client
	endpoint string
}

type UploadRequest struct {
	Domain string
	Type   string
	Cdn    string
	Time   time.Time
	Filter bool
}

const _UPLOADER_ENDPOINT = "http://jjh1267:7892"
const _UPLOADER_LINK_SERVER_ENDPOINT = "http://xs604:26001"

// http://jjh1267:7892/v2/upload/retry/upload?domain=qn-pcdngw.cdn.huya.com&hour=$(date --date="$begin" +%Y%m%d%H%M)&type=cdn&cdn=dnlivestream&filter=false
func (cli *uploaderClient) UploadRetryUpload(ctx context.Context, req *UploadRequest) error {
	URL := "http://jjh1267:7892/v2/upload/retry/upload"
	query := url.Values{}
	query.Add("domain", req.Domain)
	query.Add("hour", req.Time.Local().Format("200601021504"))
	query.Add("type", "cdn")
	query.Add("cdn", req.Cdn)
	query.Add("filter", strconv.FormatBool(req.Filter))

	URL += "?" + query.Encode()
	res, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	r, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	fmt.Printf("/v2/upload/retry/upload: '%s'\n", r)

	return nil
}

type LogLink struct {
	Size      int64
	URL       string
	Key       string
	StartTime string `json:"startTime"` // time format: time.DateTime
	EndTime   string `json:"endTime"`   // time format: time.DateTime
}

type LogListRequest struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Domain    string    `json:"domain"`
	Type      string    `json:"cdn"`
}

func (r *LogListRequest) MarshalJSON() ([]byte, error) {
	type alias LogListRequest
	tmp := struct {
		*alias
		StartTime string
		EndTime   string
	}{
		alias:     (*alias)(r),
		StartTime: r.StartTime.Format(time.DateTime),
		EndTime:   r.EndTime.Format(time.DateTime),
	}
	return json.Marshal(tmp)
}

type LogListResponse struct {
	Result struct {
		LogLinks []LogLink `json:"logs"`
	} `json:"result_desc"`
}

// curl -s -XPOST 'http://xs604:26001/v2/huyalog/list' -d '{"startTime":"2023-12-25 23:00:00","endTime":"2023-12-25 23:59:59","domain": "qn-pcdngw.cdn.huya.com", "type": "cdn"}' -H 'content-type:application/json'

func (cli *uploaderClient) HuyalogList(ctx context.Context, req *LogListRequest) (*LogListResponse, error) {
	var res LogListResponse
	var body = new(bytes.Buffer)

	json.NewEncoder(body).Encode(req)

	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, cli.endpoint+"/v2/huyalog/list", body)
	if err != nil {
		return nil, fmt.Errorf("%w: http.NewRequestWithContent: %w", ErrInvalid, err)
	}
	hreq.Header.Set("Content-Type", "application/json; utf-8")
	hres, err := cli.Do(hreq)
	if err != nil {
		return nil, fmt.Errorf("%w: http.Client.Do: %w", ErrProtocol, err)
	}
	defer hres.Body.Close()

	if err = json.NewDecoder(hres.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("%w: json.Decode: %w", ErrProtocol, err)
	}
	return &res, nil
}

type logQueryClient struct {
	*http.Client
	endpoint string
}

type DomainCdnHourRequest struct {
	Cdn    string
	Domain string
	Hour   time.Time
}

type DomainCdnHourResponse struct {
	CdnTotalFlux     map[string]metastat            `json:"cdnFlux"`
	FluxTotal        metastat                       `json:"flux"`
	PerMinuteTotal   map[string]metastat            `json:"minuteFlux"`
	PerMinuteCdnFlux map[string]map[string]metastat `json:"minuteCdnFlux"`
}

const _LOGQUERY_SERVER_ENDPOINT = "http://xs989:12344"

// curl 'http://xs989:12344/v1/log/flux/domainCdnHour?domain=qn-pcdngw.cdn.huya.com&cdn=dnlivestream&hour=2023120622' | jq .
func (cli *logQueryClient) LogFluxDomainCdnHour(ctx context.Context, req *DomainCdnHourRequest) (*DomainCdnHourResponse, error) {
	var result DomainCdnHourResponse
	var URL = fmt.Sprintf(cli.endpoint+"/v1/log/flux/domainCdnHour?domain=%[1]s&cdn=%[3]s&hour=%[2]s",
		req.Domain, req.Hour.Format("2006010215"), req.Cdn)

	res, err := cli.Get(URL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil

}

var _ CdnlogUploader = (*uploaderService)(nil)

func NewCdnlogUploader() CdnlogUploader {
	return &uploaderService{
		logQueryCli: &logQueryClient{endpoint: _LOGQUERY_SERVER_ENDPOINT, Client: http.DefaultClient},
		uploaderCli: &uploaderClient{endpoint: _UPLOADER_LINK_SERVER_ENDPOINT, Client: http.DefaultClient}}
}
