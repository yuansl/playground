// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-08-22 13:42:26

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yuansl/playground/util"
)

const _ENDPOINT_DEFAULT = "http://localhost:5140"

var fatal = util.Fatal

func Send(ctx context.Context, fields []string) error {
	payload := []struct {
		Headers map[string]any `json:"headers"`
		Body    string         `json:"body"`
	}{
		{
			Headers: map[string]any{"timestamp": time.Now().Unix(), "topic": "miku-streamd-charge"},
			Body:    strings.Join(fields, ","),
		},
	}
	var buf bytes.Buffer

	json.NewEncoder(&buf).Encode(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/miku/charge/stat", &buf)
	if err != nil {
		fatal("http.NewRequestWithContext")
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fatal("http.Client.Do:", err)
	}
	defer res.Body.Close()

	io.Copy(io.Discard, res.Body)

	return nil
}

var (
	n        int
	endpoint string
)

func parseCmdArgs() {
	flag.IntVar(&n, "n", 10_000, "number of messages to send in total")
	flag.StringVar(&endpoint, "endpoint", "http://localhost:5140", "specify flume http source endpoint")
	flag.Parse()

	if endpoint == "" {
		endpoint = _ENDPOINT_DEFAULT
	}
}

func main() {
	parseCmdArgs()

	var wg sync.WaitGroup
	var climit = make(chan struct{}, runtime.NumCPU())

	for i := 0; i < n; i++ {
		climit <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() {
				<-climit
				wg.Done()
			}()
			if err := Send(context.TODO(), []string{"yujie", strconv.Itoa(rand.Int() % 100), "Shanghai", "Qiniu Limited. co"}); err != nil {
				fatal("Send error:", err)
			}
		}()
	}
	wg.Wait()
}
