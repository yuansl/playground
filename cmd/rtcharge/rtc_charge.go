package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

//go:generate stringer -type ResolutionType -linecomment
type ResolutionType int

const (
	ResolutionAUDIO ResolutionType = iota // AUDIO
	ResolutionSD                          // SD
	ResolutionHD                          // HD
	ResolutionUHD                         // UHD
)

const (
	_480p = 480 * 480
	_720p = 1280 * 720
)

func ResolutionTypeOf(area int64) ResolutionType {
	if area <= 0 {
		return ResolutionAUDIO
	}
	if area < _480p {
		return ResolutionSD
	}
	if area < _720p {
		return ResolutionHD
	}
	return ResolutionUHD
}

type RecordTags struct {
	AppId         string `json:"appId"`
	Group         string
	Method        string
	PlayerId      string
	RoomId        string
	RoomServerId  string
	StateCenterId string `json:"stateCenterId"`
	Time          int64  `json:"time,string"`
	TrackId       string `json:"trackId"`
	Type          string
	Uid           string
	UUID          string
}

type RecordFields struct {
	Bytes    int64
	Duration int64
	Height   int64
	Width    int64
	Area     int64 `json:"-"`
}

type RtcChargeRecord struct {
	Name   string       `json:"name"`
	Fields RecordFields `json:"fields"`
	Tags   RecordTags   `json:"tags"`
}

func (r *RtcChargeRecord) String() string {
	md5sum := md5.Sum([]byte(fmt.Sprintf("%s-%s-%s-%s-%s-%s-%s-%d-%s-%s",
		r.Tags.AppId, r.Tags.Group, r.Tags.Method, r.Tags.PlayerId, r.Tags.RoomId, r.Tags.RoomServerId, r.Tags.StateCenterId, r.Tags.Time, r.Tags.TrackId, r.Tags.Type)))
	return string(md5sum[:])
}

type RtcTimeCorrectKey struct {
	Uid      string
	AppId    string
	PlayerId string
	RoomId   string
	Method   string
	TrackId  string
}

type RtcCorrectKeyRecords struct {
	Key     RtcTimeCorrectKey
	Records []RtcChargeRecord
}

type RtcChargeGroupKey struct {
	Uid      string
	AppId    string
	PlayerId string
	RoomId   string
	Time     int64
	Method   string
}

type RtcChargeGroupRecords struct {
	Key     RtcChargeGroupKey
	Records []RtcChargeRecord
}

type RtcGroupReduceCharge struct {
	RtcChargeGroupKey
	Area     int64
	Duration int64
	Bytes    int64
}

type RtcUserPlayer struct {
	Uid        string
	AppId      string
	PlayerId   string
	RoomId     string
	Resolution ResolutionType
	Method     string
	Time       int64
}

type RtcUserPlayerRecords struct {
	Key     RtcUserPlayer
	Records []RtcGroupReduceCharge
}

type RtcChargeValue struct {
	Duration int64
	Flow     int64
}

type RtcUserPlayerCharge struct {
	RtcUserPlayer
	RtcChargeValue
}

type RtcUser struct {
	Uid        string
	AppId      string
	Resolution ResolutionType
	Method     string
	Time       int64
}

type RtcUserRecords struct {
	Key     RtcUser
	Records []RtcUserPlayerCharge
}

type RtcUserCharge struct {
	RtcUser
	RtcChargeValue
}

func loadChargeRecords(filename string) []RtcChargeRecord {
	fp, err := os.Open(filename)
	if err != nil {
		fatal("os.Open:", err)
	}
	defer fp.Close()

	var records []RtcChargeRecord

	for r := bufio.NewReader(fp); ; {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fatal("bufio.ReadBytes:", err)
			}
			break
		}

		line = bytes.TrimRight(line, "\n")

		var record RtcChargeRecord

		err = json.Unmarshal(line, &record)
		if err != nil {
			fatal("json.Unmarshal:", err)
		}
		records = append(records, record)
	}
	return records
}
