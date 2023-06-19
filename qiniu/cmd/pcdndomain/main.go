package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/qbox/net-deftones/fusionrobot/traffic_syncer/domain_manager"
	"github.com/qbox/net-deftones/logger"
)

var globalOptions struct {
	mysqluri string
	from     time.Time
	to       time.Time
}

type TrafficRepository interface {
	ListDomainsWithTraffic(ctx context.Context, cdn string, from, to time.Time) ([]string, error)
	ListPcdnDomains(ctx context.Context, month time.Time) ([]string, error)
}

type trafficRepository struct {
	*sqlx.DB
	namedCdns map[string]int
}

func NewMysqlRepository(mysqluri string) TrafficRepository {
	db, err := sqlx.Open("mysql", mysqluri)
	if err != nil {
		fatalf("sqlx.Open: %v\n", err)
	}
	return &trafficRepository{DB: db, namedCdns: map[string]int{
		"pdntaiwu": 33,
	}}
}

func (repo *trafficRepository) ListDomainsWithTraffic(ctx context.Context, cdn string, from, to time.Time) ([]string, error) {
	var domains []string
	var buf strings.Builder

	fmt.Fprintf(&buf, "select distinct(domain) from raw_day_traffic_%s where day>=%s and day <%s and cdn=%d and source_type=1",
		from.Format("2006_01"), from.Format("20060102"), to.Format("20060102"), repo.namedCdns[cdn])
	err := repo.Select(&domains, buf.String())
	if err != nil {
		return nil, fmt.Errorf("repo.sqlx.Select: %v", err)
	}

	return domains, nil
}

func (repo *trafficRepository) ListPcdnDomains(ctx context.Context, month time.Time) ([]string, error) {
	var domains []string
	var buf strings.Builder

	fmt.Fprintf(&buf, "select distinct(domain) from domain_info where month=%s and product='pcdn'",
		month.Format("200601"))
	err := repo.Select(&domains, buf.String())
	if err != nil {
		return nil, fmt.Errorf("repo.sqlx.Select: %v", err)
	}

	return domains, nil
}

func fatalf(format string, v ...any) {
	fmt.Fprintf(os.Stderr, "fatal error:"+format, v...)
	os.Exit(1)
}

type DomainManager = domain_manager.PcdnDomainManager

type Set struct {
	keys map[string]struct{}
}

func (s *Set) Exists(elem string) bool {
	_, exists := s.keys[elem]
	return exists
}

func (s *Set) Append(elem string) {
	if s.Exists(elem) {
		return
	}
	s.keys[elem] = struct{}{}
}

func (s *Set) Members() []string {
	membs := []string{}
	for k := range s.keys {
		membs = append(membs, k)
	}
	return membs
}

func newSet(elems []string) *Set {
	uniq := make(map[string]struct{})
	for _, elem := range elems {
		if _, exists := uniq[elem]; exists {
			continue
		}
		uniq[elem] = struct{}{}
	}
	return &Set{keys: uniq}
}

func setdiff(set1, set2 *Set) []string {
	diffset := newSet(nil)

	for _, domain := range set1.Members() {
		if !set2.Exists(domain) {
			diffset.Append(domain)
		}
	}
	return diffset.Members()
}

func findSetdiffOfDomains(ctx context.Context, trafficsrv TrafficRepository, from, to time.Time) []string {
	diff := newSet(nil)
	domains, err := trafficsrv.ListPcdnDomains(ctx, from)
	if err != nil {
		fatalf("m.ListDomainInfos: %v", err)
	}
	set2 := newSet(domains)

	for day := from; day.Before(to); day = day.AddDate(0, 0, +1) {
		domains2, err := trafficsrv.ListDomainsWithTraffic(ctx, "pdntaiwu", day, day.AddDate(0, 0, +1))
		if err != nil {
			fatalf("traffic.ListDomains: %v", err)
		}
		for _, v := range setdiff(newSet(domains2), set2) {
			diff.Append(v)
		}
	}

	return diff.Members()
}

func parseCmdArgs() {
	flag.StringVar(&globalOptions.mysqluri, "mysql", "(127.0.0.1:3306)/traffic?parseTime=true&loc=UTC", "specify the mysql server addr")
	flag.TextVar(&globalOptions.from, "from", time.Now(), "specify start date")
	flag.TextVar(&globalOptions.to, "to", time.Now(), "specify end date")
	flag.Parse()
}

func main() {
	parseCmdArgs()

	repo := NewMysqlRepository(globalOptions.mysqluri)
	ctx := logger.NewContext(context.TODO(), logger.New())

	domains := findSetdiffOfDomains(ctx, repo, globalOptions.from, globalOptions.to)
	for _, domain := range domains {
		fmt.Println(domain)
	}
}
