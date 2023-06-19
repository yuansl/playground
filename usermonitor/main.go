package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/qbox/net-deftones/deftone_monitor/db"
	"github.com/qbox/net-deftones/deftone_monitor/submonitor"
	"github.com/qbox/net-deftones/logger"
)

type domainWarnBroker struct{}

var _ db.UserDefinedWarnBroker = domainWarnBroker{}

type UserDefinedWarn = submonitor.UserDefinedWarn
type GetWarnHistoryReq = submonitor.GetWarnHistoryReq

func (domainWarnBroker) BatchInsert(records []UserDefinedWarn) error { return nil }
func (domainWarnBroker) Insert(record *UserDefinedWarn) error        { return nil }
func (domainWarnBroker) Find(q *GetWarnHistoryReq, start, end time.Time) (warns []UserDefinedWarn, marker string, err error) {
	return
}
func (domainWarnBroker) Count(q *GetWarnHistoryReq, start, end time.Time) (count int, err error) {
	return
}

type domainRecordBroker struct{}

var _ db.UserDefinedRecordBroker = domainRecordBroker{}

func (domainRecordBroker) Record(domain, metric string, warnAt time.Time) error {
	log.Printf("Record: domain: '%s', metric: '%s', timestamp: %v\n", domain, metric, warnAt)
	return nil
}

type customNotify struct{}

var _ submonitor.Notifier = customNotify{}

func (customNotify) Notify(ctx context.Context, warning *db.UserDefinedWarn) error {
	log.Printf("Notify: warning: `%v`\n", warning)
	return nil
}

var f string
var pprof string

func init() {
	flag.StringVar(&f, "f", "", "configuration file")
	flag.StringVar(&pprof, "pprof", ":6060", "go tool pprof address")
}

func main() {
	flag.Parse()

	go func() {
		log.Println(http.ListenAndServe(pprof, nil))
	}()

	var conf submonitor.Config
	LoadConfig(&conf)

	trafficService := submonitor.NewFusionTrafficService(&conf.TrafficServiceConf)

	usermonitor := submonitor.NewUserDefinedMonitor(
		&conf.UserDefinedMonitorConf,
		submonitor.NewRuleDao(&conf.UserdefinedRuleServiceConf),
		trafficService,
		domainWarnBroker{},
		domainRecordBroker{},
		submonitor.NewDomainManager(conf.DomainMgrConf, trafficService),
		submonitor.NewCdnBossClient(&conf.CdnbossServiceConf),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	usermonitor.Check(logger.ContextWithLogger(ctx, logger.New()))
}

func LoadConfig(conf *submonitor.Config) {
	fp, err := os.Open(f)
	if err != nil {
		log.Fatal("os.Open failed:", err)
	}

	err = json.NewDecoder(fp).Decode(conf)
	if err != nil {
		log.Fatal("invalid format of configuration file, expected json format")
	}
}
