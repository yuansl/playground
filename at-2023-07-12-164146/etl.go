package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

//go:generate stringer -type DataType -linecomment
type DataType int

const (
	BandwidthData DataType = iota // bandwidth
)

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

func ListEtlTasks(r *EtlTaskRequest) (*EtlTaskResponse, error) {
	var payload EtlTaskResponse

	res, err := http.Get("http://xs201:12324/v5/etl/tasks/" + r.Id)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&payload)

	return &payload, err
}

type EtlRequest struct {
	Cdn          string    `json:"cdn"`
	Domains      string    `json:"domain"`
	Hour         time.Time `json:"-"`
	Type         DataType  `json:"type"`
	UseOnlineLog bool      `json:"useOnlineLog"`
	Force        bool      `json:"force"`
	Manual       bool      `json:"manual"`
}

func (r *EtlRequest) MarshalJSON() ([]byte, error) {
	type alias EtlRequest
	x := struct {
		*alias
		Hour string `json:"hour"`
	}{
		alias: (*alias)(r),
		Hour:  r.Hour.Format("2006-01-02-15"),
	}
	return json.Marshal(x)
}

type EtlResponse struct {
	TaskId string
}

func CreateEtlTask(r *EtlRequest) (*EtlResponse, error) {
	buf := bytes.NewBuffer(nil)

	_ = json.NewEncoder(buf).Encode(r)

	res, err := http.Post("http://xs201:12324/v5/etl/tasks", "application/json", buf)
	if err != nil {
		return nil, fmt.Errorf("http.Post: %v", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %v", err)
	}

	if res.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("%w: status=%s => %s", ErrBadRequest, res.Status, string(data))
	}
	return &EtlResponse{TaskId: string(data)}, nil
}
