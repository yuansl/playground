package main

import (
	"context"
	"fmt"
	"os"

	"github.com/qbox/net-deftones/fusionrobot/traffic_syncer/domain_manager/pcdn"
	"github.com/qbox/net-deftones/logger"
)

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

func main() {
	dm := pcdn.NewDomainManager("http://fusiondomainproxy.fusion.internal.qiniu.io",
		pcdn.WithCredential("557TpseUM8ovpfUhaw8gfa2DQ0104ZScM-BTIcBx", "d9xLPyreEG59pR01sRQcFywhm4huL-XEpHHcVa90"))
	ctx := logger.NewContext(context.Background(), logger.New())
	infos, err := dm.ListDomainInfos(ctx)
	if err != nil {
		fatal("dm.ListDomainInfos:", err)
	}
	fmt.Printf("infos: %+v\n", infos)

	info, err := dm.GetDomainInfo(ctx, "picccdl.cdn.51touxiang.com")
	if err != nil {
		fatal("dm.GetDomainInfo:", err)
	}
	fmt.Printf("info: %+v\n", info)
}
