package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type LogLink struct {
	Domain    string
	Url       string
	Size      int64
	Name      string
	Timestamp time.Time
}

type LogLinkRequest struct {
	Domains string    `json:"domains"`
	Day     time.Time `json:"-"`
}

func (r *LogLinkRequest) MarshalJSON() ([]byte, error) {
	type alias LogLinkRequest

	x := struct {
		*alias
		Day string `json:"day"`
	}{
		alias: (*alias)(r),
		Day:   r.Day.Format("2006-01-02"),
	}
	return json.Marshal(&x)
}

type LogLinkResponse struct {
	Result map[string][]LogLink `json:"data"`
}

func ListLogLinks(r *LogLinkRequest) (*LogLinkResponse, error) {
	buf := bytes.NewBuffer(nil)

	_ = json.NewEncoder(buf).Encode(r)

	res, err := http.Post("http://xs605:26001/v2/log/list", "application/json; charset=utf-8", buf)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var payload LogLinkResponse

	err = json.NewDecoder(res.Body).Decode(&payload)

	return &payload, err
}
