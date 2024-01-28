package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	awdb "github.com/godaner/awdb-golang/awdb-golang"
	"github.com/yuansl/playground/util"
)

type IPInfo struct {
	Accuracy  []byte `awdb:"accuracy"`
	Areacode  []byte `awdb:"areacode"`
	Asnumber  []byte `awdb:"asnumber"`
	City      []byte `awdb:"city"`
	Continent []byte `awdb:"continent"`
	Country   []byte `awdb:"country"`
	Isp       []byte `awdb:"isp"`
	Latwgs    []byte `awdb:"latwgs"`
	Lngwgs    []byte `awdb:"lngwgs"`
	Owner     []byte `awdb:"owner"`
	Radius    []byte `awdb:"radius"`
	Province  []byte `awdb:"province"`
	Source    []byte `awdb:"source"`
	Timezone  []byte `awdb:"timezone"`
	Zipcode   []byte `awdb:"zipcode"`
}

func (ip *IPInfo) String() string {
	return fmt.Sprintf("Accuracy: %s\nLatwgs: %s\nLngwgs: %s\nContinent: %s\nAreacode: %s\nCountry: %s\nProvince: %s\nCity: %s\nIsp: %s\nTimezone: %s\nAsnumber: %s\nOwner: %s\nSource: %s\nZipcode: %s\n",
		ip.Accuracy, ip.Latwgs, ip.Lngwgs, ip.Continent, ip.Areacode, ip.Country, ip.Province, ip.City, ip.Isp, ip.Timezone, ip.Asnumber, ip.Owner, ip.Source, ip.Zipcode)
}

var _options struct {
	ip     string
	dbfile string
}

func parseOptions() {
	flag.StringVar(&_options.ip, "ip", "", "")
	flag.StringVar(&_options.dbfile, "awdb", "ipv4.awdb", "specify ip db file in .awdb format")
	flag.Parse()

	if _options.ip == "" && len(os.Args) >= 2 {
		_options.ip = os.Args[len(os.Args)-1]
	}
}

func main() {
	parseOptions()

	reader, err := awdb.Open(_options.dbfile)
	if err != nil {
		util.Fatal("awdb.Open error:", err)
	}

	var result IPInfo
	if err = reader.Lookup(net.ParseIP(_options.ip), &result); err != nil {
		util.Fatal("awdb.Reader.LookUp():", err)
	}

	fmt.Printf("info of ip '%s':\n%s\n", _options.ip, &result)
}
