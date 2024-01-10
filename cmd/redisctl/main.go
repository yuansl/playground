// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-12-06 16:00:53

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/util/codec"
	"github.com/redis/go-redis/v9"
)

var (
	_cdnProviders string
	_redisURI     string
)

func parseCmdArgs() {
	flag.StringVar(&_cdnProviders, "cdns", "", "specify cdn provider names, seperated by comma")
	flag.StringVar(&_redisURI, "redis", "redis://xs615:6379?addr=xs616:6379&addr=xs617:6379&dial_timeout=2s&read_timeout=3s", "specify redis URI")
	flag.Parse()
}

type DomainInfo struct {
	Name     string
	Key      string
	Platform string
	Uid      uint
	Conflict bool
}

type Redis = *redis.ClusterClient

func run(ctx context.Context, rdb Redis, decoder codec.Codec) {
	for _, cdn := range strings.Split(_cdnProviders, ",") {
		if cdn == "" {
			continue
		}

		keys, err := rdb.HGetAll(ctx, fmt.Sprintf("defy:fscdn:domaindict:%s", cdn)).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			panic("rdb.HGETALL:" + err.Error())
		}
		fmt.Printf("got %d domains from cdn %q\n", len(keys), cdn)

		for key, rawdomain := range keys {
			var domain DomainInfo
			if err = decoder.Decode([]byte(rawdomain), &domain); err != nil {
				panic("decoder.Decode:" + err.Error())
			}
			if domain.Name == "" || domain.Key == "" {
				fmt.Printf("invalid form domain found: domain=%q, info='%+v' from cdn %q\n", key, domain, cdn)
			}
		}
	}
}

func main() {
	parseCmdArgs()

	options, err := redis.ParseClusterURL(_redisURI)
	if err != nil {
		panic(err)
	}
	rdb := redis.NewClusterClient(options)
	decoder := codec.NewGobCodec()
	ctx := logger.NewContext(context.TODO(), logger.New())

	run(ctx, rdb, decoder)
}
