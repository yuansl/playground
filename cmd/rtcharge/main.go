//go:build ignore

package main

import (
	"flag"
	"fmt"
	"sort"
	"time"

	"github.com/yuansl/playground/stream"
	"github.com/yuansl/playground/utils"
)

var (
	filename string
)

func parseCmdArgs() {
	flag.StringVar(&filename, "f", "", "specify json_rtc_charge filename")
	flag.Parse()
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

const _AGG_DURATION_SECONDS_DEFAULT = 300

func accumulate(records []RtcChargeRecord) {
	// for each record: set record.area = record.width * record.height
	records = stream.NewStream[RtcChargeRecord, RtcChargeRecord](records).
		Distinct(func(r RtcChargeRecord) any {
			return r.String()
		}).
		ForEach(func(r RtcChargeRecord) RtcChargeRecord {
			r.Fields.Area = r.Fields.Width * r.Fields.Height
			return r
		}).
		Collect()

	reduceCharges := stream.NewStream[RtcChargeGroupRecords, []RtcGroupReduceCharge](
		stream.GroupBy(records,
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
			Collect()).
		Aggregate(stream.Aggregator[RtcChargeGroupRecords, any, []RtcGroupReduceCharge]{
			Supplier: func() any {
				return utils.NewSet[*RtcGroupReduceCharge]()
			},
			Transform: func(z any, r RtcChargeGroupRecords) {
				set := z.(*utils.Set[*RtcGroupReduceCharge])
				charge := RtcGroupReduceCharge{RtcChargeGroupKey: r.Key}

				for j := 0; j < len(r.Records); j++ {
					rr := &r.Records[j]
					charge.Area += rr.Fields.Area
					charge.Duration = max(charge.Duration, rr.Fields.Duration)
					charge.Bytes += rr.Fields.Bytes
				}
				set.Add(&charge)
			},
			Collect: func(z any) []RtcGroupReduceCharge {
				set := z.(*utils.Set[*RtcGroupReduceCharge])
				reduceCharges := make([]RtcGroupReduceCharge, 0, set.Size())

				for i := 0; i < set.Size(); i++ {
					r := set.Get(i)

					reduceCharges = append(reduceCharges, *r)
				}
				return reduceCharges
			},
		}) // []RtcGroupReduceCharge

	playerCharge := stream.NewStream[RtcUserPlayerRecords, []RtcUserPlayerCharge](
		stream.GroupBy(reduceCharges,
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
			Collect()).
		Aggregate(stream.Aggregator[RtcUserPlayerRecords, any, []RtcUserPlayerCharge]{
			Supplier: func() any {
				return utils.NewSet[*RtcUserPlayerCharge]()
			},
			Transform: func(z any, r RtcUserPlayerRecords) {
				if len(r.Records) == 0 {
					return
				}

				set := z.(*utils.Set[*RtcUserPlayerCharge])
				charge := RtcUserPlayerCharge{
					RtcUserPlayer: r.Key,
				}

				sort.Slice(r.Records, func(i, j int) bool {
					return r.Records[i].Time < r.Records[j].Time
				})
				charge.Duration += r.Records[0].Duration
				charge.Flow += r.Records[0].Bytes
				for i := 1; i < len(r.Records); i++ {
					r0, r1 := &r.Records[i-1], &r.Records[i]

					if interval := r1.Time - r0.Time; r1.Duration > interval {
						r1.Duration = interval
					}
					charge.Duration += r1.Duration
					charge.Flow += r1.Bytes
				}
				charge.Duration = min(charge.Duration, _AGG_DURATION_SECONDS_DEFAULT)

				set.Add(&charge)
			},
			Collect: func(z any) []RtcUserPlayerCharge {
				set := z.(*utils.Set[*RtcUserPlayerCharge])
				playerCharges := make([]RtcUserPlayerCharge, 0, set.Size())

				for i := 0; i < set.Size(); i++ {
					r := set.Get(i)
					playerCharges = append(playerCharges, *r)
				}
				return playerCharges
			},
		}) // []RtcUserPlayerCharge

	PrettyPrint_PlayerCharge(playerCharge)

	userCharge := stream.NewStream[RtcUserRecords, []RtcUserCharge](
		stream.GroupBy(playerCharge,
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
			Collect()).
		Aggregate(stream.Aggregator[RtcUserRecords, any, []RtcUserCharge]{
			Supplier: func() any { return utils.NewSet[*RtcUserCharge]() },
			Transform: func(z any, r RtcUserRecords) {
				set := z.(*utils.Set[*RtcUserCharge])
				charge := RtcUserCharge{RtcUser: r.Key}

				for j := 0; j < len(r.Records); j++ {
					rr := &r.Records[j]

					charge.Duration += rr.Duration
					charge.Flow += rr.Flow
				}
				set.Add(&charge)
			},
			Collect: func(z any) []RtcUserCharge {
				set := z.(*utils.Set[*RtcUserCharge])
				userCharges := make([]RtcUserCharge, 0, set.Size())

				for i := 0; i < set.Size(); i++ {
					r := set.Get(i)

					userCharges = append(userCharges, *r)
				}
				return userCharges
			},
		}) // []RtcUserCharge

	PrettyPrint_UserCharge(userCharge)
}

func main() {
	parseCmdArgs()

	records := loadChargeRecords(filename)

	accumulate(records)
}
