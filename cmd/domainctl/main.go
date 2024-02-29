package main

import (
	"fmt"

	"github.com/qbox/net-deftones/fusionrobot"
	"github.com/qbox/net-deftones/util"
)

func main() {
	client, err := fusionrobot.NewDomainUIDClient("http://fusiondomain.fusion.internal.qiniu.io")
	if err != nil {
		util.Fatal(err)
	}

	uid, err := client.GetUid(".cdn.huya.com")
	if err != nil {
		util.Fatal(err)
	}
	fmt.Printf("uid of domain .cdn.huya.com: %d\n", uid)
}
