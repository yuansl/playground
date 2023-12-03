package main

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

var __instanceid = instanceid()

func requestGenerator() func() []byte {
	var __index uint32

	return func() []byte {
		id := [16]byte{}

		// byte 0-7
		checksum := md5.Sum(__instanceid)
		fmt.Printf("len(checksum)=%d\n", len(checksum))
		copy(id[:7], checksum[:7])
		// byte 7-12
		timestamp := time.Now().Unix()
		binary.BigEndian.PutUint32(id[7:12], uint32(timestamp))
		// byte 12-16
		binary.BigEndian.PutUint32(id[12:], atomic.AddUint32(&__index, +1))

		return id[:]
	}
}

func instanceid() []byte {
	machineid := make([]byte, 10)

	setMachineid(machineid)
	binary.BigEndian.PutUint32(machineid[6:10], uint32(os.Getpid()))
	return machineid
}

func setMachineid(machine []byte) {
	intfs, err := net.Interfaces()
	if err != nil {
		fatal("net.Interfaces() error: %v\n", err)
	}
	for _, intf := range intfs {
		if ha := intf.HardwareAddr; len(ha) > 0 {
			copy(machine, ha[:len(machine)])
			break
		}
	}
}

func main() {
	requestid := requestGenerator()

	fmt.Printf("requestid=%q\n", hex.EncodeToString(requestid()))
}

func startHttpServer() {
	var reqcounter int64
	requestid := requestGenerator()

	http.HandleFunc("/metric", func(w http.ResponseWriter, req *http.Request) {
		atomic.AddInt64(&reqcounter, +1)
		req.ParseForm()

		id := requestid()

		log.Printf("#%d: requestid(hex):%q\n", atomic.LoadInt64(&reqcounter), hex.EncodeToString(id))
	})
	http.ListenAndServe(":8080", nil)
}

func fatal(format string, v ...any) {
	fmt.Fprintf(os.Stderr, format, v...)
	os.Exit(1)
}
