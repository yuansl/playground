package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type uploaderService struct {
	endpoint string
}

type metastat struct {
	Analyze int64 `json:"analyze"`
	Upload  int64 `json:"upload"`
	Store   int64 `json:"store"`
	Log     int64 `json:"log"`
	API     int64 `json:"api"`
}

const _UPLOADER_ENDPOINT_DEFAULT = "http://xs989:12344"

// curl 'http://xs989:12344/v1/log/flux/domainCdnHour?domain=qn-pcdngw.cdn.huya.com&cdn=dnlivestream&hour=2023120622' | jq .
func (srv *uploaderService) fetchOneHourUploadStat(ctx context.Context, domain, cdn string, hour time.Time) ([]UploadStat, error) {
	var stats []UploadStat
	var result struct {
		CdnTotalFlux     map[string]metastat            `json:"cdnFlux"`
		FluxTotal        metastat                       `json:"flux"`
		PerMinuteTotal   map[string]metastat            `json:"minuteFlux"`
		PerMinuteCdnFlux map[string]map[string]metastat `json:"minuteCdnFlux"`
	}
	URL := fmt.Sprintf(srv.endpoint+"/v1/log/flux/domainCdnHour?domain=%[1]s&cdn=%[3]s&hour=%[2]s", domain, hour.Format("2006010215"), cdn)
	res, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
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

	for _, domain := range domains {
		for hour := begin; hour.Before(end); hour = hour.Add(1 * time.Hour) {
			stats0, err := srv.fetchOneHourUploadStat(ctx, domain, cdn, hour)
			if err != nil {
				return nil, err
			}
			stats = append(stats, stats0...)
		}
	}
	return stats, nil
}

type UploadRequest struct {
	Domain string
	Type   string
	Cdn    string
	Time   time.Time
}

// http://jjh1267:7892/v2/upload/retry/upload?domain=qn-pcdngw.cdn.huya.com&hour=$(date --date="$begin" +%Y%m%d%H%M)&type=cdn&cdn=dnlivestream&filter=false
func (srv *uploaderService) retryUpload(ctx context.Context, domain, cdn string, hour time.Time) error {
	URL := "http://jjh1267:7892/v2/upload/retry/upload"
	query := url.Values{}
	query.Add("domain", domain)
	query.Add("hour", hour.Local().Format("200601021504"))
	query.Add("type", "cdn")
	query.Add("cdn", cdn)
	query.Add("filter", "false")

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

func (srv *uploaderService) Retry(ctx context.Context, domains []string, cdn string, start time.Time, end time.Time) error {
	for _, domain := range domains {
		for hour := start; hour.Before(end); hour = hour.Add(1 * time.Hour) {
			if err := srv.retryUpload(ctx, domain, cdn, hour); err != nil {
				return err
			}
		}
	}
	return nil
}

var _ CdnlogUploader = (*uploaderService)(nil)

func NewCdnlogUploader() CdnlogUploader {
	return &uploaderService{endpoint: _UPLOADER_ENDPOINT_DEFAULT}
}
