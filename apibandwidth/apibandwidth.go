// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2021-08-25 15:12:18

// === Go Playground ===
// Execute the snippet with Ctl-Return
// Provide custom arguments to compile with Alt-Return
// Remove the snippet completely with its dir and all files M-x `go-playground-rm`

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	filename  string
	c         int
	cdn       string
	start     string
	end       string
	batchSize int
)

func init() {
	flag.StringVar(&filename, "f", "", "specify filename of a domain list")
	flag.IntVar(&c, "c", runtime.NumCPU(), "concurrency level")
	flag.StringVar(&cdn, "cdn", "aliyun", "specify cdn provider name")
	flag.StringVar(&start, "start", "2021-08-23", "specify start date")
	flag.StringVar(&end, "end", "2021-08-25", "specify end date")
	flag.IntVar(&batchSize, "batch", 10, "batch size of domains per request")
}

func main() {
	flag.Parse()

	fp, err := os.Open(filename)
	if err != nil {
		log.Fatal("os.Open: ", err)
	}
	defer fp.Close()

	from, err := time.Parse("2006-01-02", start)
	if err != nil {
		log.Fatal("time.Parse(start) error: ", err)
	}
	to, err := time.Parse("2006-01-02", end)
	if err != nil {
		log.Fatal("time.Parse(end) error: ", err)
	}

	reader := bufio.NewReader(fp)

	var wg sync.WaitGroup
	var climit = make(chan struct{}, c)
	var domains []string

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Fatal("bufio.ReadBytes: ", err)
			}

			refreshDomainBandwidth(cdn, domains, from, to)
			break
		}
		domain := strings.TrimSpace(string(line))
		if domain == "domain" {
			continue
		}

		domains = append(domains, domain)
		if len(domains) > batchSize {
			domains2 := make([]string, len(domains))
			copy(domains2, domains)

			climit <- struct{}{}
			wg.Add(1)
			go func(cdn string, domains []string) {
				defer func() {
					<-climit
					wg.Done()
				}()

				fmt.Printf("processing domains %v at [%s,%s]\n", domains, from, to)

				refreshDomainBandwidth(cdn, domains, from, to)

			}(cdn, domains2)

			domains = domains[:0]
		}
	}
	wg.Wait()
}

func refreshDomainBandwidth(cdn string, domains []string, from, to time.Time) {
	var q = url.Values{}
	q.Set("cdn", cdn)
	q.Set("domains", strings.Join(domains, ","))
	q.Set("from", from.Format("2006-01-02"))
	q.Set("to", to.Format("2006-01-02"))

	resp, err := http.Get("http://xs321:18214/v1/cdn/time/range/bandwidth?" + q.Encode())
	if err != nil {
		log.Fatal("http.Get error: ", err)
	}
	defer resp.Body.Close()

	io.Copy(io.Discard, resp.Body)
}
