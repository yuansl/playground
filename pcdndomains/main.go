package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/qbox/net-deftones/fusionrobot/traffic_syncer"
)

var addr string
var timeoutsec int

func init() {
	flag.StringVar(&addr, "addr", "", "specify address of pcdn domains manager")
	flag.IntVar(&timeoutsec, "timeout", 10, "timeout in seconds")
}

func main() {
	flag.Parse()

	pcdn := traffic_syncer.NewPcdnDomainManager(addr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutsec)*time.Second)
	defer cancel()

	domains, err := pcdn.ListDomainInfos(ctx)
	if err != nil {
		log.Fatal("pcdn.ListDOmainInfos: ", err)
	}

	for _, domain := range domains {
		log.Printf("domain: %+v\n", domain)
	}
}
