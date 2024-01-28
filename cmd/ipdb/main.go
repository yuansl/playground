// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-12-15 15:20:25

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ipipdotnet/ipdb-go"
	"github.com/yuansl/playground/util"
)

const LANGUAGE = "CN"

var _options struct {
	ipdb    string
	verbose bool
}

func parseCmdOptions() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <ip>\n", os.Args[0])
		os.Exit(1)
	}

	flag.StringVar(&_options.ipdb, "ipdb", "neo.ipv4.ipdb", "specify ipdb file")
	flag.BoolVar(&_options.verbose, "v", false, "verbose")
	flag.Parse()
}

func describeipdb(db *ipdb.City) {
	// db.Reload("/path/to/city.ipv4.ipdb") // 更新 ipdb 文件后可调用 Reload 方法重新加载内容
	fmt.Println("ipdb info:")
	fmt.Println("\tipv4 enabled: ", db.IsIPv4())  // check database support ip type
	fmt.Println("\tipv6 enabled: ", db.IsIPv6())  // check database support ip type
	fmt.Println("\tbuild time: ", db.BuildTime()) // database build time
	fmt.Println("\tlanguage: ", db.Languages())   // database support language
	if _options.verbose {
		fmt.Println("\tfields: ", db.Fields()) // database support fields
	}
}

func main() {
	parseCmdOptions()

	ip := os.Args[len(os.Args)-1]
	db, err := ipdb.NewCity(_options.ipdb)
	if err != nil {
		util.Fatal(err)
	}
	
	describeipdb(db)

	ipdb.NewDistrict(name string)
	// fmt.Println(db.Find("1.1.1.1", "CN"))            // return []string
	// fmt.Println(db.FindMap("118.28.8.8", "CN"))      // return map[string]string

	info, err := db.FindInfo(ip, LANGUAGE)
	if err != nil {
		util.Fatal(err)
	}

	fmt.Printf("\ninfo of ip '%s':\n", ip)

	fmt.Printf("Area code: %s\nContinent: %s\nCountry: %s(%s)\nCity: %s\nIDC: %s\n",
		info.AreaCode, info.ContinentCode, info.CountryName, info.CountryCode, info.CityName, info.IDC)
}
