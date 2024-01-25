package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/qbox/net-deftones/util"
)

var fatal = util.Fatal

var _options struct {
	addr string
}

func parseCmdOptions() {
	flag.StringVar(&_options.addr, "addr", ":5140", "listen address of the server")
	flag.Parse()
}

func main() {
	parseCmdOptions()

	queue := make(chan []Message, runtime.NumCPU())
	go func() {
		fp, err := os.OpenFile("/tmp/syslog.2", os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0664)
		if err != nil {
			fatal("os.OpenFile:", err)
		}
		defer fp.Close()

		for msgs := range queue {
			for _, msg := range msgs {
				fmt.Fprintf(fp, "%s\n", msg.Body)
			}
		}
	}()
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		var msgs []Message
		err := json.NewDecoder(req.Body).Decode(&msgs)

		if err != nil {
			fatal("json.Decode:", err)
		}
		fmt.Printf("content-length: %s, msgs.len=%d\n", req.Header.Get("Content-Length"), len(msgs))
		if len(msgs) > 1 {
			for i := 0; i < len(msgs); i++ {
				fmt.Printf("msg.header: %+v\n", msgs[i].Headers)
			}
		}
		queue <- msgs
	})
	println(http.ListenAndServe(_options.addr, nil))
}
