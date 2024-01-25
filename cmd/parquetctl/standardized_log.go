package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/parquet-go/parquet-go"
	"github.com/yuansl/playground/util"
)

type StandardizedLog struct {
	CdnProvider           string            `parquet:"cdn_provider" db:"cdn_provider"`
	ClientIp              string            `parquet:"client_ip" db:"client_ip"`
	ContentType           string            `parquet:"content_type" db:"content_type"`
	Domain                string            `parquet:"domain" db:"domain"`
	Url                   string            `parquet:"url" db:"url"`
	RequestTime           string            `parquet:"request_time" db:"request_time"`
	ResponseTime          string            `parquet:"response_time" db:"response_time"`
	ServerIp              string            `parquet:"server_ip" db:"server_ip"`
	RequestMethod         string            `parquet:"request_method" db:"request_method"`
	Scheme                string            `parquet:"scheme" db:"scheme"`
	ServerProtocol        string            `parquet:"server_protocol" db:"server_protocol"`
	StatusCode            string            `parquet:"status_code" db:"status_code"`
	HttpRange             string            `parquet:"http_range" db:"http_range"`
	BytesSent             string            `parquet:"bytes_sent" db:"bytes_sent"`
	BodyBytesSent         string            `parquet:"body_bytes_sent" db:"body_bytes_sent"`
	Hitmiss               string            `parquet:"hitmiss" db:"hitmiss"`
	HttpReferer           string            `parquet:"http_referer" db:"http_referer"`
	Ua                    string            `parquet:"ua" db:"ua"`
	ServerPort            string            `parquet:"server_port" db:"server_port"`
	FirstByteTime         string            `parquet:"first_byte_time" db:"first_byte_time"`
	HttpXForwardFor       string            `parquet:"http_x_forward_for" db:"http_x_forward_for"`
	RequestLength         string            `parquet:"request_length" db:"request_length"`
	RequestId             string            `parquet:"request_id" db:"request_id"`
	SentHttpContentLength string            `parquet:"sent_http_content_length" db:"sent_http_content_length"`
	RequestBodyLength     string            `parquet:"request_body_length" db:"request_body_length"`
	UpstreamResponseTime  string            `parquet:"upstream_response_time" db:"upstream_response_time"`
	HttpCookie            string            `parquet:"http_cookie" db:"http_cookie"`
	Upstream5xx           string            `parquet:"upstream_5xx" db:"upstream_5xx"`
	Delay                 int64             `parquet:"delay" db:"delay"`
	ServerRegion          string            `parquet:"server_region" db:"server_region"`
	PcdnCdnBytesSent      string            `parquet:"pcdn_cdn_bytes_sent" db:"pcdn_cdn_bytes_sent"`
	PdnFluxType           string            `parquet:"pdn_flux_type" db:"pdn_flux_type"`
	BT                    string            `parquet:"_bt" db:"_bt"`
	Extras                map[string]string `parquet:"extras" db:"-"`
}

var _parquetSchema = parquet.SchemaOf((*StandardizedLog)(nil))

func readParquetFile(filename string, ch chan<- *StandardizedLog) {
	fp, err := os.Open(filename)
	if err != nil {
		util.Fatal("os.Open:", err)
	}
	defer fp.Close()

	r := parquet.NewReader(fp, _parquetSchema)
	defer r.Close()

	start := time.Now()

	var counter atomic.Int32
	for {
		var stdlog StandardizedLog

		err := r.Read(&stdlog)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fmt.Printf("parquet.Read error: %v\n", err)
			}
			break
		}
		ch <- &stdlog

		counter.Add(+1)
		// fmt.Printf("stdlog: %+v\n", stdlog)
	}
	fmt.Printf("read %d lines from file %s in %v\n", counter.Load(), filename, time.Since(start))
}

func SaveasParquet(w io.Writer, ch <-chan *StandardizedLog) {
	pw := parquet.NewWriter(w, _parquetSchema)
	defer pw.Close()
	var counter atomic.Int32

	for stdlog := range ch {
		err := pw.Write(stdlog)
		if err != nil {
			util.Fatal("parquet.Write:", err)
		}
		counter.Add(+1)
	}

	fmt.Printf("write %d lines total\n", counter.Load())
}

func mergeParquetFiles(filenames []string) error {
	fp, err := os.OpenFile("merge.parquet", os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0660)
	if err != nil {
		return fmt.Errorf("os.OpenFile(merge.parquet): %w", err)
	}
	defer fp.Close()

	var wg sync.WaitGroup
	var ch = make(chan *StandardizedLog)
	var done = make(chan bool)

	go func() {
		w := bufio.NewWriterSize(fp, BUFSIZE)

		SaveasParquet(w, ch)
		w.Flush()

		done <- true
	}()
	for _, filename := range filenames {
		wg.Add(1)
		go func(filename string) {
			defer wg.Done()
			readParquetFile(filename, ch)
		}(filename)
	}
	wg.Wait()
	close(ch)
	<-done
	return nil
}
