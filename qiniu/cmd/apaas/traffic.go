package main

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/qbox/net-deftones/util"
)

var ErrDatabase = errors.New("database: internal error")

type DataType int

const (
	DataTypeBandwidth DataType = iota
	DataTypeDynBandwidth
	DataTypeReqcount
	DataTypeDynReqcount
)

type Region int

const (
	RegionChina Region = iota + 1
	RegionForeign
	_
	RegionAmeu
	RegionAsia
	RegionSea
	RegionSA
	RegionOC
	_
	RegionNozone
)

type TrafficStat struct {
	Timeseries []int64
	Day        time.Time
}

type TrafficService interface {
	GetCdnTrafficOf(ctx context.Context, from, to time.Time, dataType DataType, uid uint32) ([]TrafficStat, error)
	GetAllOfCdnTraffics(ctx context.Context, from, to time.Time, dataType DataType) ([]TrafficStat, error)
}

type trafficService struct {
	db *sqlx.DB
}

func (srv *trafficService) getCdnTrafficByMonth(ctx context.Context, from time.Time, to time.Time, dataType DataType, uid uint32) ([]TrafficStat, error) {
	rawsql := "select day, a.domain,region,slot0,slot1,slot2,slot3,slot4,slot5,slot6,slot7,slot8,slot9,slot10,slot11,slot12,slot13,slot14,slot15,slot16,slot17,slot18,slot19,slot20,slot21,slot22,slot23,slot24,slot25,slot26,slot27,slot28,slot29,slot30,slot31,slot32,slot33,slot34,slot35,slot36,slot37,slot38,slot39,slot40,slot41,slot42,slot43,slot44,slot45,slot46,slot47,slot48,slot49,slot50,slot51,slot52,slot53,slot54,slot55,slot56,slot57,slot58,slot59,slot60,slot61,slot62,slot63,slot64,slot65,slot66,slot67,slot68,slot69,slot70,slot71,slot72,slot73,slot74,slot75,slot76,slot77,slot78,slot79,slot80,slot81,slot82,slot83,slot84,slot85,slot86,slot87,slot88,slot89,slot90,slot91,slot92,slot93,slot94,slot95,slot96,slot97,slot98,slot99,slot100,slot101,slot102,slot103,slot104,slot105,slot106,slot107,slot108,slot109,slot110,slot111,slot112,slot113,slot114,slot115,slot116,slot117,slot118,slot119,slot120,slot121,slot122,slot123,slot124,slot125,slot126,slot127,slot128,slot129,slot130,slot131,slot132,slot133,slot134,slot135,slot136,slot137,slot138,slot139,slot140,slot141,slot142,slot143,slot144,slot145,slot146,slot147,slot148,slot149,slot150,slot151,slot152,slot153,slot154,slot155,slot156,slot157,slot158,slot159,slot160,slot161,slot162,slot163,slot164,slot165,slot166,slot167,slot168,slot169,slot170,slot171,slot172,slot173,slot174,slot175,slot176,slot177,slot178,slot179,slot180,slot181,slot182,slot183,slot184,slot185,slot186,slot187,slot188,slot189,slot190,slot191,slot192,slot193,slot194,slot195,slot196,slot197,slot198,slot199,slot200,slot201,slot202,slot203,slot204,slot205,slot206,slot207,slot208,slot209,slot210,slot211,slot212,slot213,slot214,slot215,slot216,slot217,slot218,slot219,slot220,slot221,slot222,slot223,slot224,slot225,slot226,slot227,slot228,slot229,slot230,slot231,slot232,slot233,slot234,slot235,slot236,slot237,slot238,slot239,slot240,slot241,slot242,slot243,slot244,slot245,slot246,slot247,slot248,slot249,slot250,slot251,slot252,slot253,slot254,slot255,slot256,slot257,slot258,slot259,slot260,slot261,slot262,slot263,slot264,slot265,slot266,slot267,slot268,slot269,slot270,slot271,slot272,slot273,slot274,slot275,slot276,slot277,slot278,slot279,slot280,slot281,slot282,slot283,slot284,slot285,slot286,slot287 from domain_day_traffic_" + from.Format("2006_01") + " a join domain_info b on a.domain=b.domain where a.day >= :from and a.day <:to and b.month=:month and b.product='fusion' and a.data_type=:data_type and b.uid=:uid"
	args := map[string]any{
		"from":      from.Format("20060102"),
		"to":        to.Format("20060102"),
		"month":     from.Format("200601"),
		"data_type": dataType,
		"uid":       uid,
	}
	rows, err := srv.db.NamedQueryContext(ctx, rawsql, args)
	if err != nil {
		return nil, fmt.Errorf("%w: db.NamedQuery: %v", ErrDatabase, err)
	}
	defer rows.Close()

	var traffics []TrafficStat

	for rows.Next() {
		var row struct {
			Day    int64  `db:"day"`
			Domain string `db:"domain"`
			Region Region `db:"region"`
			Slots
		}

		err = rows.StructScan(&row)
		if err != nil {
			return nil, fmt.Errorf("%w: StructScan: %v", ErrDatabase, err)
		}
		traffics = append(traffics, TrafficStat{
			Day:        time.Date(int(row.Day)/10000, time.Month(row.Day)/100000%100, int(row.Day)%100, 0, 0, 0, 0, time.Local),
			Timeseries: row.toTimeseries(),
		})
	}
	return traffics, nil
}

func (srv *trafficService) GetCdnTrafficOf(ctx context.Context, from time.Time, to time.Time, dataType DataType, uid uint32) ([]TrafficStat, error) {
	var traffics []TrafficStat

	for t := from; t.Before(to); t = t.AddDate(0, +1, 0) {
		traffics0, err := srv.getCdnTrafficByMonth(ctx, t, t.AddDate(0, +1, 0), dataType, uid)
		if err != nil {
			return nil, fmt.Errorf("srv.getCdnTrafficByMonth: %v", err)
		}
		traffics = append(traffics, traffics0...)
	}
	return traffics, nil
}

func (srv *trafficService) getAllCdnTrafficsByMonth(ctx context.Context, from time.Time, to time.Time, dataType DataType) ([]TrafficStat, error) {

	rows, err := srv.db.NamedQueryContext(ctx,
		"select b.uid,day,a.domain,region,slot0,slot1,slot2,slot3,slot4,slot5,slot6,slot7,slot8,slot9,slot10,slot11,slot12,slot13,slot14,slot15,slot16,slot17,slot18,slot19,slot20,slot21,slot22,slot23,slot24,slot25,slot26,slot27,slot28,slot29,slot30,slot31,slot32,slot33,slot34,slot35,slot36,slot37,slot38,slot39,slot40,slot41,slot42,slot43,slot44,slot45,slot46,slot47,slot48,slot49,slot50,slot51,slot52,slot53,slot54,slot55,slot56,slot57,slot58,slot59,slot60,slot61,slot62,slot63,slot64,slot65,slot66,slot67,slot68,slot69,slot70,slot71,slot72,slot73,slot74,slot75,slot76,slot77,slot78,slot79,slot80,slot81,slot82,slot83,slot84,slot85,slot86,slot87,slot88,slot89,slot90,slot91,slot92,slot93,slot94,slot95,slot96,slot97,slot98,slot99,slot100,slot101,slot102,slot103,slot104,slot105,slot106,slot107,slot108,slot109,slot110,slot111,slot112,slot113,slot114,slot115,slot116,slot117,slot118,slot119,slot120,slot121,slot122,slot123,slot124,slot125,slot126,slot127,slot128,slot129,slot130,slot131,slot132,slot133,slot134,slot135,slot136,slot137,slot138,slot139,slot140,slot141,slot142,slot143,slot144,slot145,slot146,slot147,slot148,slot149,slot150,slot151,slot152,slot153,slot154,slot155,slot156,slot157,slot158,slot159,slot160,slot161,slot162,slot163,slot164,slot165,slot166,slot167,slot168,slot169,slot170,slot171,slot172,slot173,slot174,slot175,slot176,slot177,slot178,slot179,slot180,slot181,slot182,slot183,slot184,slot185,slot186,slot187,slot188,slot189,slot190,slot191,slot192,slot193,slot194,slot195,slot196,slot197,slot198,slot199,slot200,slot201,slot202,slot203,slot204,slot205,slot206,slot207,slot208,slot209,slot210,slot211,slot212,slot213,slot214,slot215,slot216,slot217,slot218,slot219,slot220,slot221,slot222,slot223,slot224,slot225,slot226,slot227,slot228,slot229,slot230,slot231,slot232,slot233,slot234,slot235,slot236,slot237,slot238,slot239,slot240,slot241,slot242,slot243,slot244,slot245,slot246,slot247,slot248,slot249,slot250,slot251,slot252,slot253,slot254,slot255,slot256,slot257,slot258,slot259,slot260,slot261,slot262,slot263,slot264,slot265,slot266,slot267,slot268,slot269,slot270,slot271,slot272,slot273,slot274,slot275,slot276,slot277,slot278,slot279,slot280,slot281,slot282,slot283,slot284,slot285,slot286,slot287 from domain_day_traffic_"+from.Format("2006_01")+" a join domain_info b on a.domain=b.domain where a.day >= :from and a.day <:to and b.month=:month and b.product='fusion' and a.data_type=:data_type", map[string]any{
			"from":      from.Format("20060102"),
			"to":        to.Format("20060102"),
			"month":     from.Format("200601"),
			"data_type": dataType,
		})
	if err != nil {
		return nil, fmt.Errorf("%w: db.NamedQuery: %v", ErrDatabase, err)
	}
	defer rows.Close()

	var traffics []TrafficStat

	for rows.Next() {
		var row struct {
			Uid    uint32 `db:"uid"`
			Day    int64  `db:"day"`
			Domain string `db:"domain"`
			Region Region `db:"region"`
			Slots
		}
		err = rows.StructScan(&row)
		if err != nil {
			return nil, fmt.Errorf("%w: StructScan: %v", ErrDatabase, err)
		}

		traffics = append(traffics, TrafficStat{
			Day:        time.Date(int(row.Day)/10000, time.Month(row.Day)/100000%100, int(row.Day)%100, 0, 0, 0, 0, time.Local),
			Timeseries: row.toTimeseries(),
		})
	}
	return traffics, nil
}

func (srv *trafficService) GetAllOfCdnTraffics(ctx context.Context, from, to time.Time, dataType DataType) ([]TrafficStat, error) {
	var traffics []TrafficStat

	for t := from; t.Before(to); t = t.AddDate(0, +1, 0) {
		traffics0, err := srv.getAllCdnTrafficsByMonth(ctx, t, t.AddDate(0, +1, 0), dataType)
		if err != nil {
			return nil, fmt.Errorf("srv.getCdnTrafficByMonth: %v", err)
		}
		traffics = append(traffics, traffics0...)
	}
	return traffics, nil
}

type TrafficServiceOption = util.Option

type trafficServiceOptions struct {
	maxLifeTime time.Duration
	maxConns    int
	DSN         string
}

func WithMaxConns(maxConns int) TrafficServiceOption {
	return util.OptionFunc(func(opt any) {
		opt.(*trafficServiceOptions).maxConns = maxConns
	})
}

func WithDSN(dsn string) TrafficServiceOption {
	return util.OptionFunc(func(opt any) {
		opt.(*trafficServiceOptions).DSN = dsn
	})
}

func initTrafficServiceOptions(opts ...TrafficServiceOption) *trafficServiceOptions {
	var options trafficServiceOptions

	for _, op := range opts {
		op.Apply(&options)
	}
	if options.DSN == "" {
		options.DSN = "traffic_admin:admin@/traffic?parseTime=true&loc=Local"
	}
	if options.maxLifeTime <= 0 {
		options.maxLifeTime = 2 * time.Minute
	}
	if options.maxConns <= 0 {
		options.maxConns = runtime.NumCPU()
	}

	return &options
}

func NewTrafficService(opts ...TrafficServiceOption) TrafficService {
	options := initTrafficServiceOptions(opts...)

	db, err := sqlx.Connect("mysql", options.DSN)
	if err != nil {
		fatal("sqlx.Open:", err)
	}
	db.SetConnMaxLifetime(options.maxLifeTime)
	db.SetMaxIdleConns(options.maxConns)
	db.SetMaxOpenConns(options.maxConns)

	return &trafficService{db: db}
}
