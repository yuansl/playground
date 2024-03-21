// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-03-06 22:18:14

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/parquet-go/parquet-go"
	"github.com/qbox/net-deftones/util"
)

type CdnEdgeLog struct {
	Domain      string `parquet:"domain"`
	RequestTime int64  `parquet:"request_time"`
}

type Codec[T any] interface {
	Encode(v T) ([]byte, error)
	Decode(data []byte, v *T) error
}

type cdnlogCodec[T any] struct {
	bufferPool sync.Pool
}

// Decode implements Codec.
func (c *cdnlogCodec[T]) Decode(_ []byte, v *T) error {
	panic("unimplemented")
}

// Encode implements Codec.
func (c *cdnlogCodec[T]) Encode(v T) ([]byte, error) {
	buf := c.bufferPool.Get().(*bytes.Buffer)
	defer c.bufferPool.Put(buf)
	buf.Reset()

	w := parquet.NewGenericWriter[T](buf)
	if _, err := w.Write([]T{v}); err != nil {
		return nil, err
	}
	w.Close()

	return buf.Bytes(), nil
}

func newcdnlogCodec[T any]() *cdnlogCodec[T] {
	return &cdnlogCodec[T]{
		bufferPool: sync.Pool{New: func() any { return bytes.NewBuffer(nil) }},
	}
}

var _ Codec[*CdnEdgeLog] = (*cdnlogCodec[*CdnEdgeLog])(nil)

func main() {
	logs := []CdnEdgeLog{{Domain: "www.1.example.com", RequestTime: time.Now().UnixNano()}, {Domain: "www.2.example.com", RequestTime: time.Now().UnixNano()}}

	c := newcdnlogCodec[*CdnEdgeLog]()
	bytes, err := c.Encode(&CdnEdgeLog{Domain: "www.1.example.com", RequestTime: time.Now().UnixNano()})
	if err != nil {
		util.Fatal("c.Encode:", err)
	}
	fmt.Printf("output.Bytes.len: %d, %d\n", len(bytes), bytes)

	buf := parquet.NewGenericBuffer[CdnEdgeLog]()
	n, err := buf.Write(logs)
	if err != nil {
		util.Fatal("Buffer.Write:", err)
	}
	fmt.Printf("Write %d rows to parquet buffer: %d\n", n, buf.Len())

	var rows [10]parquet.Row
	n, err = buf.Rows().ReadRows(rows[:])
	if err != nil {
		if errors.Is(err, io.EOF) {
			if n > 0 {

			}
		}
		util.Fatal("ReadRows:", err)
	}
	fmt.Printf("Read %d rows from buffer: %+v\n", n, rows)
	for i, row := range rows[:n] {
		fmt.Printf("#%d row: %s\n", i, row)
	}
}
