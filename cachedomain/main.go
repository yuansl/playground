package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/qbox/net-deftones/fscdn.v2"
	"github.com/qbox/net-deftones/logger"
)

var timeoutsec int

func init() {
	flag.IntVar(&timeoutsec, "timeout", 10, "specify timeout in seconds")
}

func main() {
	flag.Parse()

	timeout := time.Duration(timeoutsec) * time.Second

	cache := fscdn.NewDomainInfoRedisCache([]string{"10.200.20.54:6379"}, timeout)

	ctx, cancel := context.WithTimeout(logger.ContextWithLogger(context.Background(), logger.New()), timeout)
	defer cancel()

	cdn := "somecdn"

	log.Println("cache.Set ...")

	cache.Set(ctx, cdn, []fscdn.DomainInfo{
		{Domain: "www.example.com", Platform: "web", Key: "123"},
		{Domain: "www.2.example.com", Platform: "web", Key: "123"},
	})

	log.Println("cache.Get ...")
	res, err := cache.Get(ctx, cdn)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			log.Println("redis: empty list or set")

			os.Exit(0)
		}
		log.Fatal("cache.Get error: ", err)
	}

	log.Println("resust: ", res)
}
