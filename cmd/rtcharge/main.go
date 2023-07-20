// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-07-19 15:57:36

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"golang.org/x/exp/constraints"

	"github.com/yuansl/playground/stream"
	"github.com/yuansl/playground/utils"
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

func max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

var (
	filename string
)

func parseCmdArgs() {
	flag.StringVar(&filename, "f", "", "specify json_rtc_charge filename")
	flag.Parse()
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

func distinct(records []RtcChargeRecord) []RtcChargeRecord {
	type UniqueKey struct {
		RecordFields
		RecordTags
	}
	var uniq = make(map[UniqueKey][]RtcChargeRecord)
	for _, r := range records {
		k := UniqueKey{RecordTags: r.Tags, RecordFields: r.Fields}
		uniq[k] = append(uniq[k], r)
	}
	records2 := make([]RtcChargeRecord, 0, len(uniq))

	for _, rs := range uniq {
		records2 = append(records2, rs...)
	}
	return records2
}

func PrettyPrint_UserCharge(userCharge []RtcUserCharge) {
	fmt.Println("\n\nrtc user charge:")

	sort.Slice(userCharge, func(i, j int) bool { return userCharge[i].Time < userCharge[j].Time })
	for _, r := range userCharge {
		fmt.Printf("%s: %+v\n", time.Unix(r.Time, 0).Format(time.DateTime), r)
	}
}

func PrettyPrint_PlayerCharge(playerCharge []RtcUserPlayerCharge) {
	fmt.Println("rtc player charge:")

	sort.Slice(playerCharge, func(i, j int) bool { return playerCharge[i].Time < playerCharge[j].Time })
	for _, c := range playerCharge {
		fmt.Printf("%s: %+v\n", time.Unix(c.Time, 0).Format(time.DateTime), c)
	}

}

func PrettyPrint_UserPlayerRecords(userPlayerRecords []RtcUserPlayerRecords) {
	for _, r := range userPlayerRecords {
		fmt.Printf("%+v:\n", r.Key)

		sort.Slice(r.Records, func(i, j int) bool { return r.Records[i].Time < r.Records[j].Time })

		for _, rr := range r.Records {
			fmt.Printf("%s: %+v\n", time.Unix(rr.Time, 0).Format(time.DateTime), rr)
		}
	}
}

func main() {
	parseCmdArgs()

	records := loadChargeRecords(filename)

	records = distinct(records)

	// for each record: set record.area = record.width * record.height
	records = stream.NewStream[RtcChargeRecord, RtcChargeRecord](records).
		Filter(func(r RtcChargeRecord) bool {
			return r.Tags.Method != "publish"
		}).
		Filter(func(r RtcChargeRecord) bool {
			return r.Tags.PlayerId == "cfe9825ad19545f4bd814a107906f3f4" && (1688382300 <= r.Tags.Time && r.Tags.Time < 1688382600)
		}).
		ForEach(func(r RtcChargeRecord) RtcChargeRecord {
			r.Fields.Area = r.Fields.Width * r.Fields.Height
			return r
		}).
		Set()

	groupRecords := stream.GroupBy[RtcChargeRecord, RtcChargeGroupRecords, RtcChargeGroupKey](records,
		func(r RtcChargeRecord) RtcChargeGroupKey {
			return RtcChargeGroupKey{
				Uid:      r.Tags.Uid,
				AppId:    r.Tags.AppId,
				PlayerId: r.Tags.PlayerId,
				RoomId:   r.Tags.RoomId,
				Time:     r.Tags.Time,
				Method:   r.Tags.Method,
			}
		},
		func(k RtcChargeGroupKey, records []RtcChargeRecord) RtcChargeGroupRecords {
			return RtcChargeGroupRecords{Key: k, Records: records}
		}).
		Set() // [(RtcChargeGroupKey, []RtcChargeRecord)]

	reduceCharges := stream.NewStream[RtcChargeGroupRecords, []RtcGroupReduceCharge](groupRecords).
		Collect(stream.Collector[RtcChargeGroupRecords, any, []RtcGroupReduceCharge]{
			Supplier: func() any {
				return utils.NewSet[*RtcChargeGroupRecords]()
			},
			BiConsumer: func(z any, r RtcChargeGroupRecords) {
				set := z.(*utils.Set[*RtcChargeGroupRecords])
				set.Add(&r)
			},
			Function: func(z any) []RtcGroupReduceCharge {
				set := z.(*utils.Set[*RtcChargeGroupRecords])
				reduceCharges := make([]RtcGroupReduceCharge, 0, set.Size())

				for i := 0; i < set.Size(); i++ {
					r := set.Get(i)
					charge := RtcGroupReduceCharge{
						RtcChargeGroupKey: r.Key,
					}

					for j := 0; j < len(r.Records); j++ {
						rr := &r.Records[j]

						charge.Area += rr.Fields.Area
						charge.Duration = max(charge.Duration, rr.Fields.Duration)
						charge.Bytes += rr.Fields.Bytes
					}
					reduceCharges = append(reduceCharges, charge)
				}
				return reduceCharges
			},
		}) // []RtcGroupReduceCharge

	userPlayerRecords := stream.GroupBy[RtcGroupReduceCharge, RtcUserPlayerRecords, RtcUserPlayer](reduceCharges,
		func(v RtcGroupReduceCharge) RtcUserPlayer {
			return RtcUserPlayer{
				Uid:        v.Uid,
				AppId:      v.AppId,
				PlayerId:   v.PlayerId,
				RoomId:     v.RoomId,
				Resolution: ResolutionTypeOf(v.Area),
				Method:     v.Method,
				Time:       v.Time / 300 * 300,
			}
		},
		func(k RtcUserPlayer, rs []RtcGroupReduceCharge) RtcUserPlayerRecords {
			return RtcUserPlayerRecords{Key: k, Records: rs}
		}).
		Set() // []{RtcUserPlayer, []RtcGroupReduceCharge}

	PrettyPrint_UserPlayerRecords(userPlayerRecords)

	playerCharge := stream.NewStream[RtcUserPlayerRecords, []RtcUserPlayerCharge](userPlayerRecords).
		Collect(stream.Collector[RtcUserPlayerRecords, any, []RtcUserPlayerCharge]{
			Supplier: func() any {
				return utils.NewSet[*RtcUserPlayerRecords]()
			},
			BiConsumer: func(z any, r RtcUserPlayerRecords) {
				set := z.(*utils.Set[*RtcUserPlayerRecords])
				set.Add(&r)
			},
			Function: func(z any) []RtcUserPlayerCharge {
				set := z.(*utils.Set[*RtcUserPlayerRecords])
				playerCharges := make([]RtcUserPlayerCharge, 0, set.Size())

				for i := 0; i < set.Size(); i++ {
					r := set.Get(i)
					charge := RtcUserPlayerCharge{
						RtcUserPlayer: r.Key,
					}

					for j := 0; j < len(r.Records); j++ {
						rr := &r.Records[j]

						charge.Duration += rr.Duration
						charge.Flow += rr.Bytes
					}
					playerCharges = append(playerCharges, charge)
				}
				return playerCharges
			},
		}) // []RtcUserPlayerCharge

	PrettyPrint_PlayerCharge(playerCharge)

	userRecords := stream.GroupBy[RtcUserPlayerCharge, RtcUserRecords, RtcUser](playerCharge,
		func(v RtcUserPlayerCharge) RtcUser {
			return RtcUser{
				Uid:        v.Uid,
				AppId:      v.AppId,
				Resolution: v.Resolution,
				Method:     v.Method,
				Time:       v.Time,
			}
		},
		func(k RtcUser, rs []RtcUserPlayerCharge) RtcUserRecords {
			return RtcUserRecords{Key: k, Records: rs}
		}).
		Set() // []{RtcUser, []RtcUserPlayerCharge}

	userCharge := stream.NewStream[RtcUserRecords, []RtcUserCharge](userRecords).
		Collect(stream.Collector[RtcUserRecords, any, []RtcUserCharge]{
			Supplier: func() any { return utils.NewSet[*RtcUserRecords]() },
			BiConsumer: func(z any, r RtcUserRecords) {
				set := z.(*utils.Set[*RtcUserRecords])
				set.Add(&r)
			},
			Function: func(z any) []RtcUserCharge {
				set := z.(*utils.Set[*RtcUserRecords])
				userCharges := make([]RtcUserCharge, 0, set.Size())

				for i := 0; i < set.Size(); i++ {
					r := set.Get(i)
					charge := RtcUserCharge{RtcUser: r.Key}

					for j := 0; j < len(r.Records); j++ {
						rr := &r.Records[j]

						charge.Duration += rr.Duration
						charge.Flow += rr.Flow
					}
					userCharges = append(userCharges, charge)
				}
				return userCharges
			},
		}) // []RtcUserCharge

	PrettyPrint_UserCharge(userCharge)
}
