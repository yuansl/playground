package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/parquet-go/parquet-go"
	"github.com/qbox/net-deftones/util"
)

func saveAsParquet(logs []CdnEdgeLog, filename string) error {
	fp, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		util.Fatal("os.OpenFile:", err)
	}
	w0 := bufio.NewWriterSize(fp, 32<<10)
	w := parquet.NewGenericWriter[CdnEdgeLog](w0)
	n, err := w.Write(logs)
	if err != nil {
		util.Fatal("parquet.Write:", err)
	}
	fmt.Printf("Wrote %d rows into parquet file\n", n)
	if err := w.Close(); err != nil {
		util.Fatal("parquet.Close:", err)
	}
	w0.Flush()
	return fp.Close()
}

type CdnLogParquetReader struct {
	*parquet.GenericReader[CdnEdgeLog]
	err error
}

func (r *CdnLogParquetReader) Read(rows []CdnEdgeLog) (int, error) {
	if r.err != nil {
		return -1, r.err
	}
	n, err := r.GenericReader.Read(rows)
	if err != nil {
		switch {
		case errors.Is(err, io.EOF):
			if n > 0 {
				r.err = io.EOF
				return n, nil
			}
		default:
			return -1, fmt.Errorf("parquet.GenericReader.Read: %w", err)
		}
	}
	return n, nil
}

func readFromParquet(filename string) error {
	fp2, err := os.Open(filename)
	if err != nil {
		util.Fatal("os.Open:", err)
	}
	defer fp2.Close()

	off, err := fp2.Seek(0, io.SeekStart)
	fmt.Printf("offset; %d\n", off)

	r := CdnLogParquetReader{GenericReader: parquet.NewGenericReader[CdnEdgeLog](fp2)}

	rows := [10]CdnEdgeLog{}
	n, err := r.Read(rows[:])
	if err != nil {
		util.Fatal("parquet.GenericReader.Read:", err, n)
	}
	fmt.Printf("Read %d rows from parquet file: %+v\n", n, rows[:n])

	return nil
}

func Example_parquet_file_rw() {
	logs := []CdnEdgeLog{
		{Domain: "www.1.example.com", RequestTime: time.Now().UnixNano()},
		{Domain: "www.2.example.com", RequestTime: time.Now().UnixNano()},
	}
	filename := "/tmp/some.parquet"

	if err := saveAsParquet(logs, filename); err != nil {
		util.Fatal("save to parquet:", err)
	}
	if err := readFromParquet(filename); err != nil {
		util.Fatal("Read from parquet", err)
	}

	logs = []CdnEdgeLog{
		{Domain: "www.3.example.com", RequestTime: time.Now().UnixNano()},
		{Domain: "www.2.example.com", RequestTime: time.Now().UnixNano()},
	}
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Domain <= logs[j].Domain
	})
	fmt.Printf("logs=%+v\n", logs)
}
