// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-12-15 16:10:47

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/yuansl/playground/util"
)

const MAX_FILE_SIZE = 1 << 30

const METADATA_LEN_NR_BYTES = 4

const (
	NODE_OFFSET_FAR = 80
	NODE_OFFSET_MAX = 96
)

const (
	NODE_INDEX_BASE_FAR  = 1
	NODE_INDEX_BASE_NEAR = 0
)

const (
	IPV4_BITS = net.IPv4len * 8
	IPV6_BITS = net.IPv6len * 8
)

const IP_NODE_SIZE_NR_BYTES = 2

const LANGUAGE_CN = "CN"

var (
	ErrNotFound = errors.New("ipdb: not found")
	ErrDatabase = errors.New("ipdb: database error")
)

type MetaData struct {
	Build     int64          `json:"build"`
	IPVersion uint16         `json:"ip_version"`
	Languages map[string]int `json:"languages"`
	NodeCount int64          `json:"node_count"`
	TotalSize int64          `json:"total_size"`
	Fields    []string       `json:"fields"`
}

type ReadWriter struct{}

func bitscount(ip net.IP) int {
	if ipv4 := ip.To4(); ipv4 != nil {
		return IPV4_BITS
	} else {
		return IPV6_BITS
	}
}

type IPdb struct {
	meta    *MetaData
	ipv4off int
	ipdata  []byte
}

func (db *IPdb) findNode(node, index int) int {
	off := node*8 + index*4
	nextnode := int(binary.BigEndian.Uint32(db.ipdata[off : off+4]))

	fmt.Printf("node=%d,index=%d,off=%d,nextnode=%d,next4bytes=%d\n", node, index, off, nextnode, binary.BigEndian.Uint32(db.ipdata[off+4:off+8]))

	return nextnode
}

func (db *IPdb) Search(ip net.IP) error {
	bits := bitscount(ip)
	node := 0

	if bits == IPV4_BITS {
		node = db.ipv4off // ip is a ipv4
	}
	for i := 0; i < bits && node <= int(db.meta.NodeCount); i++ {
		node = db.findNode(node, ((0xFF&int(ip[i>>3]))>>uint(7-(i%8)))&1)
	}
	fmt.Printf("got node=%d, nodecount=%d\n", node, db.meta.NodeCount)
	if node > int(db.meta.NodeCount) {
		ipNodeOffset := node - int(db.meta.NodeCount) + int(db.meta.NodeCount)*8
		ipNodeSize := int(binary.BigEndian.Uint16(db.ipdata[ipNodeOffset : ipNodeOffset+IP_NODE_SIZE_NR_BYTES]))
		if (ipNodeOffset + IP_NODE_SIZE_NR_BYTES + ipNodeSize) > len(db.ipdata) {
			return ErrDatabase
		}
		ipNode := db.ipdata[ipNodeOffset+IP_NODE_SIZE_NR_BYTES : ipNodeOffset+IP_NODE_SIZE_NR_BYTES+ipNodeSize]
		fields := bytes.Split(ipNode, []byte("\t"))
		langoff := db.meta.Languages[LANGUAGE_CN]
		ipinfos := fields[langoff : langoff+len(db.meta.Fields)]
		for i, ipinfo := range ipinfos {
			fmt.Printf("%s: %s\n", db.meta.Fields[i], ipinfo)
		}
		return nil
	}
	return ErrNotFound
}

func open(filename string) *IPdb {
	var metadata MetaData

	fp, err := os.Open(filename)
	if err != nil {
		util.Fatal(err)
	}
	dbdata, err := io.ReadAll(io.LimitReader(fp, MAX_FILE_SIZE))
	if err != nil {
		util.Fatal(err)
	}

	metalen := binary.BigEndian.Uint32(dbdata[:METADATA_LEN_NR_BYTES])
	if err = json.Unmarshal(dbdata[METADATA_LEN_NR_BYTES:METADATA_LEN_NR_BYTES+metalen], &metadata); err != nil {
		util.Fatal(err)
	}
	ipdb := &IPdb{meta: &metadata, ipdata: dbdata[METADATA_LEN_NR_BYTES+metalen:]}

	fmt.Printf("meta=%+v, len(dbdata)=%d\n", metadata, len(ipdb.ipdata))

	node := 0
	for i := 0; i < NODE_OFFSET_MAX && node < int(metadata.NodeCount); i++ {
		if i >= NODE_OFFSET_FAR {
			node = ipdb.findNode(node, NODE_INDEX_BASE_FAR)
		} else {
			node = ipdb.findNode(node, NODE_INDEX_BASE_NEAR)
		}
	}
	fmt.Printf("ipv4off=%d\n", node)
	ipdb.ipv4off = node
	return ipdb
}

var _options struct {
	ip string
}

func parseCmdOptions() {
	flag.StringVar(&_options.ip, "ip", "1.1.1.1", "specify a ip(v4|v6) address you want to query")
	flag.Parse()
}

func main() {
	parseCmdOptions()
	db := open(filepath.Join(os.Getenv("HOME"), "/Downloads/neo.ipv4.ipdb"))

	err := db.Search(net.ParseIP(_options.ip))
	if err != nil {
		util.Fatal(err)
	}
}
