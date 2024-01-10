package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"

	"github.com/colinmarc/hdfs"
	"github.com/parquet-go/parquet-go"
	"golang.org/x/sync/errgroup"

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
	dir      string
	namenode string
}

func parseCmdOptions() {
	// "kodofs://defy-dw/offlinelogs/day=20231112/hour=19/_domain=qn-pcdngw.cdn.huya.com/cdn=dnlivestream/"
	flag.StringVar(&_options.dir, "dir", "/", "specify directory name for traversaling")
	// "jjh1274:8020"
	flag.StringVar(&_options.namenode, "namenode", "localhost:8020", "namenode of hdfs")
	flag.Parse()
}

func readParquet(r io.ReaderAt) error {
	pr := parquet.NewReader(r, parquet.SchemaOf(new(StandardizedLog)))
	flux := int64(0)
	nlines := int64(0)

	for {
		var stdlog StandardizedLog

		err := pr.Read(&stdlog)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("parquest.Read: %w", err)
			}
			break
		}
		bytesSent, err := strconv.Atoi(stdlog.BytesSent)
		if err != nil {
			return fmt.Errorf("ill-formed bytessent: %q", stdlog.BytesSent)
		}
		flux += int64(bytesSent)
		nlines++
	}
	fmt.Printf("got %d bytes sent, processed %d log lines in total\n", flux, nlines)

	return nil
}

func main() {
	parseCmdOptions()

	fs, err := hdfs.New(_options.namenode)
	if err != nil {
		util.Fatal(err)
	}
	dirents, err := fs.ReadDir(_options.dir)
	if err != nil {
		util.Fatal(err)
	}

	var wg errgroup.Group

	for _, ent := range dirents {
		if ent.Mode().IsRegular() {
			_ent := ent
			wg.Go(func() error {
				reader, err := fs.Open(_ent.Name())
				if err != nil {
					util.Fatal("fs.Open:", err)
				}

				fmt.Printf("processing file: %q ...\n", _ent.Name())
				return readParquet(reader)
			})
		}
	}
	wg.Wait()
}
