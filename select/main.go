package main

import (
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

var num int
var concurrency int

func init() {
	flag.IntVar(&num, "n", 1000_000, "specify number of test")
	flag.IntVar(&concurrency, "c", runtime.NumCPU(), "specify concurrency limit")
	flag.Parse()
}

func main() {
	compareBigendian()
}

func compareBigendian() {
	testcases := [...]int{0x0102, 0x030405, 0x05060708, 0x0809101112}

	for _, num := range testcases {
		var x []byte

		b := *(*byte)(unsafe.Pointer(&num))
		var payload [8]byte
		switch {
		case num <= math.MaxInt8:
			payload[0] = byte(num)
			x = bigendian(uint8(num))
		case num <= math.MaxInt16:
			binary.BigEndian.PutUint16(payload[:], uint16(num))
			x = bigendian(uint16(num))
		case num <= math.MaxInt32:
			x = bigendian(uint32(num))
			binary.BigEndian.PutUint32(payload[:], uint32(num))
		default:
			x = bigendian(uint64(num))
			binary.BigEndian.PutUint64(payload[:], uint64(num))
		}

		fmt.Printf("%#010x/b[0]=%#0x in bigendian: %q vs %q/payload[0]=%#0x\n",
			num, b, hex.EncodeToString(x), hex.EncodeToString(payload[:]), payload[0])
	}
}

func sendHttpRequestInParallel() {
	var wg sync.WaitGroup
	var climit = make(chan struct{}, concurrency)

	wg.Add(num)
	for i := 0; i < num; i++ {
		climit <- struct{}{}
		go func() {
			defer func() {
				<-climit
				wg.Done()
			}()
			resp, err := http.Post("http://localhost:8080/metric", "application/octet-stream", nil)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			io.Copy(io.Discard, resp.Body)
		}()
	}
	wg.Wait()
}

type usize interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64
}

const (
	SizeOfint8 = 1 << iota
	Sizeofint16
	SizeOfint32
	SizeOfint64
)

func bigendian[T usize](num T) []byte {
	var payload [8]byte

	switch sizeof := unsafe.Sizeof(num); sizeof {
	case SizeOfint8:
		payload[0] = byte(num)

	case Sizeofint16:
		payload[0] = byte(uint16(num) >> 8)
		payload[1] = byte(num)

	case SizeOfint32:
		payload[0] = byte(uint32(num) >> 24)
		payload[1] = byte(uint32(num) >> 16)
		payload[2] = byte(uint32(num) >> 8)
		payload[3] = byte(num)

	case SizeOfint64:
		payload[0] = byte(uint64(num) >> 56)
		payload[1] = byte(uint64(num) >> 48)
		payload[2] = byte(uint64(num) >> 40)
		payload[3] = byte(uint64(num) >> 32)
		payload[4] = byte(uint64(num) >> 24)
		payload[5] = byte(uint64(num) >> 16)
		payload[6] = byte(uint64(num) >> 8)
		payload[7] = byte(num)

	default:
		panic(fmt.Sprintf("unsupport sizeof: %d", sizeof))
	}
	return payload[:]
}

func nocaseSelect() {
	go func() {
		for range time.Tick(1 * time.Second) {
			fmt.Println("Tick")
		}
	}()
	select {}
}
