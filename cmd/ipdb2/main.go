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

	"github.com/qbox/net-deftones/util"
)

const MAX_DBFILE_SIZE = 1 << 30

const IPDB_METADATA_LEN_NR_BYTES = 4

const (
	IPDB_NODE_OFFSET_FAR = 80
	IPDB_NODE_OFFSET_MAX = 96
)

const (
	IPDB_NODE_INDEX_BASE_FAR  = 1
	IPDB_NODE_INDEX_BASE_NEAR = 0
)

const IPDB_NODE_SIZE_NR_BYTES = 2

const (
	IPV4_BITS = net.IPv4len * 8
	IPV6_BITS = net.IPv6len * 8
)

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

var _ipinfoFields = [...]string{
	"country_name",
	"region_name",
	"city_name",
	"owner_domain",
	"isp_domain",
	"latitude",
	"longitude",
	"timezone",
	"utc_offset",
	"china_admin_code",
	"idd_code",
	"country_code",
	"continent_code",
}

type IPInfo struct {
	Latwgs    []byte `awdb:"latwgs" ipdb:"latitude"`
	Lngwgs    []byte `awdb:"lngwgs" ipdb:"longitude"`
	Continent []byte `awdb:"continent" ipdb:"continent_code"`
	Areacode  []byte `awdb:"areacode" ipdb:"country_code"`
	Country   []byte `awdb:"country" ipdb:"country_name"`
	City      []byte `awdb:"city" ipdb:"city_name"`
	Accuracy  []byte `awdb:"accuracy"`
	Asnumber  []byte `awdb:"asnumber"`
	Isp       []byte `awdb:"isp" ipdb:"isp_domain"`
	Owner     []byte `awdb:"owner" ipdb:"owner_domain"`
	Radius    []byte `awdb:"radius"`
	Province  []byte `awdb:"province" ipdb:"region_name"`
	Source    []byte `awdb:"source"`
	Timezone  []byte `awdb:"timezone" ipdb:"utc_offset"`
	Zipcode   []byte `awdb:"zipcode" ipdb:"idd_code"`
}

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

type DB struct {
	Header struct {
		metasize uint32
		meta     []byte // in json
	}
	nodes []struct {
		zero uint32
		one  uint32
	}
	leafs []struct {
		size    uint16
		content []byte
	}
}

func (db *IPdb) findNode(node, bit int) int {
	off := node*8 + bit*4
	nextnode := int(binary.BigEndian.Uint32(db.ipdata[off : off+4]))

	if node >= 96 {
		fmt.Printf("node=%d,binbit=%d,off=%d,nextnode=%d(%#[4]x)\n", node, bit, off, nextnode)
	}

	return nextnode
}

func (db *IPdb) putNode(node *int, index int) {
	off := (*node)*8 + index*4
	(*node)++
	fmt.Printf("put node %d at %d, index=%d\n", *node, off, index)
	binary.BigEndian.PutUint32(db.ipdata[off:off+4], uint32(*node))
}

func (db *IPdb) nodeOffsetof(ip net.IP) int {
	node := 0
	nbits := bitscount(ip)

	if nbits == IPV4_BITS {
		node = db.ipv4off // ip is a ipv4
		ip = ip.To4()
	}
	ip2 := binary.BigEndian.Uint32(ip)
	fmt.Printf("nbits of ip %s(%#08x): %d\n", ip, ip2, nbits)

	for i := 0; i < nbits && node <= int(db.meta.NodeCount); i++ {
		bit := ((0xff & ip[i/8]) >> (7 - (i % 8))) & 0x01

		node = db.findNode(node, int(bit))
		fmt.Printf("ip[%d]=%#08b(%#02[2]x,bit=%#02x),node=%d\n", i, ip[i/8], bit, node)
	}
	fmt.Printf("node=%d\n", node)
	if node > int(db.meta.NodeCount) {
		return int(db.meta.NodeCount)*8 + node - int(db.meta.NodeCount)
	}
	return -1
}

func (db *IPdb) Put(ip net.IP) {

}

func (db *IPdb) Search(ip net.IP) error {
	ipNodeOffset := db.nodeOffsetof(ip)

	if ipNodeOffset == -1 {
		return ErrNotFound
	}
	// |---2 Byte node-size-length----|------ (node-size-length bytes) node info data----|
	ipNodeSize := int(binary.BigEndian.Uint16(db.ipdata[ipNodeOffset : ipNodeOffset+IPDB_NODE_SIZE_NR_BYTES]))

	if (ipNodeOffset + IPDB_NODE_SIZE_NR_BYTES + ipNodeSize) > len(db.ipdata) {
		return ErrDatabase
	}
	ipNode := db.ipdata[ipNodeOffset+IPDB_NODE_SIZE_NR_BYTES : ipNodeOffset+IPDB_NODE_SIZE_NR_BYTES+ipNodeSize]

	fmt.Printf("nodeoffset=%d,nodesize=%d,nodedata=%s\n", ipNodeOffset, ipNodeSize, ipNode)

	fields := bytes.Split(ipNode, []byte("\t"))

	langoff := db.meta.Languages[LANGUAGE_CN]
	ipinfos := fields[langoff : langoff+len(db.meta.Fields)]

	for i, ipinfo := range ipinfos {
		fmt.Printf("%s: %s\n", db.meta.Fields[i], ipinfo)
	}
	return nil
}

func (db *IPdb) init() {
	for off, node := 0, 0; off < IPDB_NODE_OFFSET_MAX; off++ {
		bit := IPDB_NODE_INDEX_BASE_NEAR
		if off >= IPDB_NODE_OFFSET_FAR {
			bit = IPDB_NODE_INDEX_BASE_FAR
		}
		db.putNode(&node, bit)
	}
}

func open(filename string) *IPdb {
	var metadata MetaData

	fp, err := os.Open(filename)
	if err != nil {
		util.Fatal(err)
	}

	dbdata, err := io.ReadAll(io.LimitReader(fp, MAX_DBFILE_SIZE))
	if err != nil {
		util.Fatal(err)
	}
	metalen := binary.BigEndian.Uint32(dbdata[:IPDB_METADATA_LEN_NR_BYTES])

	if err = json.Unmarshal(dbdata[IPDB_METADATA_LEN_NR_BYTES:IPDB_METADATA_LEN_NR_BYTES+metalen], &metadata); err != nil {
		util.Fatal(err)
	}
	fmt.Printf("meta: %+v\n", metadata)

	ipdb := &IPdb{meta: &metadata, ipdata: dbdata[IPDB_METADATA_LEN_NR_BYTES+metalen:]}
	node := 0
	for i := 0; i < IPDB_NODE_OFFSET_MAX && node < int(metadata.NodeCount); i++ {
		bit := IPDB_NODE_INDEX_BASE_NEAR
		if i >= IPDB_NODE_OFFSET_FAR {
			bit = IPDB_NODE_INDEX_BASE_FAR
		}
		node = ipdb.findNode(node, bit)
	}
	ipdb.ipv4off = node

	fmt.Printf("ipv4off start at:%d\n", node)

	return ipdb
}

var _options struct {
	ip string
	db string
}

func parseOptions() {
	flag.StringVar(&_options.ip, "ip", "1.1.1.1", "specify a ip(v4|v6) address you want to query")
	flag.StringVar(&_options.db, "db", "ipv4.ipdb", "specify ipdb file")
	flag.Parse()
}

func main() {
	parseOptions()

	db := open(_options.db)

	// for ip := 0xffff00; ip < 1000_00000; ip += 255 {
	// 	var buf [4]byte

	// 	binary.BigEndian.PutUint32(buf[:], uint32(ip))
	// 	_ip := netip.AddrFrom4(netip.AddrFrom4(buf).As4()).String()

	// 	ip := net.ParseIP("::0")
	// 	if ipv4 := ip.To4(); ipv4 != nil {
	// 		ip = ipv4
	// 	}
	// 	db.Search(ip)
	// }

	ip := net.ParseIP("::0")
	if ipv4 := ip.To4(); ipv4 != nil {
		ip = ipv4
	}
	db.Search(ip)

	// node := 0
	// bits := bitcount(ip)

	// for i := 0; i < bits; i++ {
	// 	index := (0xff & ip[i/8]) >> (7 - (i % 8)) & 0x01
	// 	db.putNode(&node, int(index))
	// 	fmt.Printf("put: ip[%d]=%d(bin:%#8[2]b),index=%d,node=%d\n", i/8, ip[i/8], index, node)
	// }
	// fmt.Printf("node=%d\n", node)
}
