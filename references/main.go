package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	_ "net/http/pprof"

	"github.com/qbox/net-deftones/fscdn.v2"
	"github.com/qbox/net-deftones/logger"
)

var refaddress string
var address string
var redisaddr string
var pprof string

func init() {
	flag.StringVar(&refaddress, "reference_address", "http://cs50:18120", "specify address of name-reference service for google cdn")
	flag.StringVar(&address, "address", ":18130", "specify listen address of name-reference service for google cdn")
	flag.StringVar(&redisaddr, "redisaddr", "cs50:6379", "sepecify address of a availabled redis server")
	flag.StringVar(&pprof, "pprof", ":6060", "specify go tool pprof listen address")
}

func main() {
	flag.Parse()

	go func() {
		http.ListenAndServe(pprof, nil)
	}()

	mapper := fscdn.NewDomainMapper(&fscdn.FusionSophonConfig{
		Address: refaddress,
		RedisConf: fscdn.RedisConfig{
			Addrs: []string{redisaddr},
		},
		DomainKeyCacheExpiry: "5m",
	}, "googlecloudcdn")

	http.HandleFunc("/v1/reference/all", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		log := logger.New()

		data, _ := httputil.DumpRequest(req, false)
		log.Infof("Request (raw header): %v\n", string(data))

		ctx := logger.ContextWithLogger(req.Context(), log)

		refs, err := mapper.GetNameReferences(ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		data, err = json.Marshal(refs)
		w.Write(data)
	})

	http.HandleFunc("/v1/reference/domain", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		ctx := logger.ContextWithLogger(req.Context(), logger.New())

		req.ParseForm()
		key := req.Form.Get("domain")
		domain, err := mapper.GetDomainByKey(ctx, key)
		if err != nil {
			w.Write([]byte(`{"error": "not found"}`))
		} else {
			w.Write([]byte(fmt.Sprintf(`{"name": "%s", "key": "%s"}`, domain, key)))
		}
	})

	log.Println(http.ListenAndServe(address, nil))
}
