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
	"fmt"
	"log"

	"github.com/ipipdotnet/ipdb-go"
)

func main() {
	db, err := ipdb.NewCity("/path/to/city.ipv4.ipdb")
	if err != nil {
		log.Fatal(err)
	}

	db.Reload("/path/to/city.ipv4.ipdb") // 更新 ipdb 文件后可调用 Reload 方法重新加载内容

	fmt.Println(db.IsIPv4())    // check database support ip type
	fmt.Println(db.IsIPv6())    // check database support ip type
	fmt.Println(db.BuildTime()) // database build time
	fmt.Println(db.Languages()) // database support language
	fmt.Println(db.Fields())    // database support fields

	fmt.Println(db.FindInfo("2001:250:200::", "CN")) // return CityInfo
	fmt.Println(db.Find("1.1.1.1", "CN"))            // return []string
	fmt.Println(db.FindMap("118.28.8.8", "CN"))      // return map[string]string
	fmt.Println(db.FindInfo("127.0.0.1", "CN"))      // return CityInfo

	fmt.Println()
}
