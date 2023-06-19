package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
)

var addr string

func parseCmdArgs() {
	flag.StringVar(&addr, "addr", ":5140", "listen address of the server")
	flag.Parse()
}

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

type Message struct {
	Headers map[string]any
	Body    string
}

func main() {
	parseCmdArgs()

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
	http.ListenAndServe(addr, nil)
}
