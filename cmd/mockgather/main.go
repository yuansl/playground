package main

import (
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/qbox/net-deftones/clients/gather"
	"github.com/qbox/net-deftones/logger"
)

var options struct {
	addr string
}

func parseOptions() {
	flag.StringVar(&options.addr, "addr", ":9090", "")
	flag.Parse()
	if options.addr == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -addr <listen addresss>\n", os.Args[0])
		os.Exit(1)
	}
}

func main() {
	parseOptions()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v2/logverse/gather", func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		if id := req.Header.Get("X-Reqid"); id != "" {
			ctx = logger.NewContext(ctx, logger.NewWith(id))
		} else {
			ctx = logger.NewContext(ctx, logger.New())
		}
		var body io.Reader = req.Body
		if compress := req.Header.Get("Content-Encoding"); strings.Contains(compress, "gzip") {
			gz, err := gzip.NewReader(body)
			if err != nil {
				logger.FromContext(ctx).Infof("gzip.NewReader: %w", err)
				return
			}
			defer gz.Close()
			body = gz
		}
		data, _ := io.ReadAll(body)
		logger.FromContext(ctx).Infof("Request(raw): body size: `%d`\n", len(data))

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(&gather.GatherResponse{Result: "OK"})
	})
	fmt.Println(http.ListenAndServe(options.addr, mux))
}
