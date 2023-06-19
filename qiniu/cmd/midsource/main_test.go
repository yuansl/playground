package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/qbox/net-deftones/deftone_monitor/db"
	"github.com/qbox/net-deftones/fscdn.v2"
	"github.com/qbox/net-deftones/logger"
)

type some struct{}

func fatal(format string, v ...any) {
	fmt.Fprintf(os.Stderr, "fatal "+format, v...)
	os.Exit(1)
}

func main() {
	midsource := db.NewMidsrcMetricAPI("http://midsource.defy.internal.qiniu.io")
	ctx := logger.ContextWithLogger(context.TODO(), logger.New())
	date := time.Date(2022, 9, 23, 0, 0, 0, 0, time.Local)
	domains, err := midsource.TopDomainsOf(ctx, fscdn.CDNProviderVolcengine, date)
	if err != nil {
		fatal("TopDomainsOf error: %v\n", err)
	}
	uniq := make(map[string]some)
	for _, domain := range domains {
		if _, exists := uniq[domain]; !exists {
			uniq[domain] = some{}
		}
	}
	fmt.Printf("top %d midsource domains:", len(uniq))
	for domain := range uniq {
		fmt.Printf("%s\n", domain)
	}
	const target = "kflow.wpscdn.cn"
	if _, exists := uniq[target]; exists {
		fmt.Printf("the domain %s exists in the midsource's TopDomainsList\n", target)
	} else {
		fmt.Printf("the domain %s doesn't exist in the midsource's TopDomainsList\n", target)
	}
}
