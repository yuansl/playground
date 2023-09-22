package loglinks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

const _LOGLINK_ENDPOINT = "http://xs605:26001"

type Client struct{ *http.Client }

type LogLink struct {
	Domain    string
	Url       string
	Size      int64
	Name      string
	Timestamp time.Time
}

type LogListRequest struct {
	Domains string    `json:"domains"`
	Day     time.Time `json:"-"`
}

func (r *LogListRequest) MarshalJSON() ([]byte, error) {
	type alias LogListRequest

	x := struct {
		*alias
		Day string `json:"day"`
	}{
		alias: (*alias)(r),
		Day:   r.Day.Format(time.DateOnly),
	}
	return json.Marshal(&x)
}

type LogListResponse struct {
	Result map[string][]LogLink `json:"data"`
}

func (client *Client) ListLogLinks(r *LogListRequest) (*LogListResponse, error) {
	buf := bytes.NewBuffer(nil)

	_ = json.NewEncoder(buf).Encode(r)

	res, err := http.Post(_LOGLINK_ENDPOINT+"/v2/log/list", "application/json; charset=utf-8", buf)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var payload LogListResponse

	err = json.NewDecoder(res.Body).Decode(&payload)

	return &payload, err
}

func NewClient() *Client {
	return &Client{}
}
