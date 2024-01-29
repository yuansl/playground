package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	awdb "github.com/godaner/awdb-golang/awdb-golang"
	"github.com/yuansl/playground/util"
)

var _options struct {
	ip     string
	dbfile string
}

func parseOptions() {
	flag.StringVar(&_options.ip, "ip", "", "")
	flag.StringVar(&_options.dbfile, "awdb", "ipv4.awdb", "specify ip db file in .awdb format")
	flag.Parse()
}

type IPInfo struct {
	Latwgs    []byte `awdb:"latwgs"`
	Lngwgs    []byte `awdb:"lngwgs"`
	Continent []byte `awdb:"continent"`
	Areacode  []byte `awdb:"areacode"`
	Country   []byte `awdb:"country"`
	City      []byte `awdb:"city"`
	Accuracy  []byte `awdb:"accuracy"`
	Asnumber  []byte `awdb:"asnumber"`
	Isp       []byte `awdb:"isp"`
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
	{
		// var result IPInfo
		// if err = reader.Lookup(net.ParseIP(_options.ip), &result); err != nil {
		// 	util.Fatal("awdb.Reader.LookUp():", err)
		// }

		fp, err := os.OpenFile("/tmp/ipdb", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			util.Fatal(err)
		}
		defer fp.Close()

		for nets := reader.Networks(); nets.Next(); {
			var res IPInfo

			net, err := nets.Network(&res)
			if err != nil {
				fmt.Fprintf(os.Stderr, "nets.Network error: %v, skip\n", err)
				var res2 map[string]any
				net, err = nets.Network(&res2)
				if err != nil {
					util.Fatal(err)
				}
				fmt.Printf("info of net(%s): %s\n", net, res2)
				fmt.Fprintf(fp, "info of net(%s): %s\n", net, &res2)
				continue
			}
			fmt.Fprintf(fp, "info of net(%s): %s\n", net, &res)
		}

		// fmt.Printf("Info of ip '%s':\n%s\n", _options.ip, &result)
	}
}
