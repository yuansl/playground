// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-07-14 13:05:34

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/parquet-go/parquet-go"
	"github.com/yuansl/playground/util"
)

type StandardizedLog struct {
	CdnProvider           string            `parquet:"cdn_provider"`
	ClientIp              string            `parquet:"client_ip"`
	ContentType           string            `parquet:"content_type"`
	Domain                string            `parquet:"domain"`
	Url                   string            `parquet:"url"`
	RequestTime           string            `parquet:"request_time"`
	ResponseTime          string            `parquet:"response_time"`
	ServerIp              string            `parquet:"server_ip"`
	RequestMethod         string            `parquet:"request_method"`
	Scheme                string            `parquet:"scheme"`
	ServerProtocol        string            `parquet:"server_protocol"`
	StatusCode            string            `parquet:"status_code"`
	HttpRange             string            `parquet:"http_range"`
	BytesSent             string            `parquet:"bytes_sent"`
	BodyBytesSent         string            `parquet:"body_bytes_sent"`
	Hitmiss               string            `parquet:"hitmiss"`
	HttpReferer           string            `parquet:"http_referer"`
	Ua                    string            `parquet:"ua"`
	ServerPort            string            `parquet:"server_port"`
	FirstByteTime         string            `parquet:"first_byte_time"`
	HttpXForwardFor       string            `parquet:"http_x_forward_for"`
	RequestLength         string            `parquet:"request_length"`
	RequestId             string            `parquet:"request_id"`
	SentHttpContentLength string            `parquet:"sent_http_content_length"`
	RequestBodyLength     string            `parquet:"request_body_length"`
	UpstreamResponseTime  string            `parquet:"upstream_response_time"`
	HttpCookie            string            `parquet:"http_cookie"`
	Upstream5xx           string            `parquet:"upstream_5xx"`
	Delay                 int64             `parquet:"delay"`
	ServerRegion          string            `parquet:"server_region"`
	PcdnCdnBytesSent      string            `parquet:"pcdn_cdn_bytes_sent"`
	PdnFluxType           string            `parquet:"pdn_flux_type"`
	Extras                map[string]string `parquet:"extras"`
}

var _options struct {
	filename string
}

func parseCmdOptions() {
	flag.StringVar(&_options.filename, "f", "", "specify parquet filename")
	flag.Parse()
}

var _parquetSchema = parquet.SchemaOf((*StandardizedLog)(nil))

func readParquetFile(filename string) {
	fp, err := os.Open(filename)
	if err != nil {
		util.Fatal("os.Open:", err)
	}
	defer fp.Close()

	r := parquet.NewReader(fp, _parquetSchema)
	defer r.Close()

	nrRows := r.NumRows()

	fmt.Printf("number rows: %v\n", nrRows)

	for {
		var stdlog StandardizedLog

		err := r.Read(&stdlog)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fmt.Printf("parquet.Read error: %v\n", err)
			}
			break
		}
		fmt.Printf("stdlog: %+v\n", stdlog)
	}
}

func writeParquet(w0 io.Writer) {
	w := parquet.NewWriter(w0, _parquetSchema)
	defer w.Close()

	err := w.Write(&StandardizedLog{
		Extras:    map[string]string{"hello, parquet": "wolrd"},
		Delay:     12434,
		HttpRange: "bytes=0-",
	})
	if err != nil {
		util.Fatal("parquet.Write:", err)
	}
	w.Close()
}

func writeFile(filename string) {
	fp, err := os.OpenFile(_options.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		util.Fatal(err)
	}
	defer fp.Close()

	writeParquet(bufio.NewWriter(fp))
}

func main() {
	parseCmdOptions()
	// writeFile(_options.filename)
	readParquetFile(_options.filename)
}
