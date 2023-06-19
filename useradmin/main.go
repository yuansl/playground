package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

var (
	uid uint
)

func init() {
	flag.UintVar(&uid, "uid", 1381044116, "uid of user")
}

func main() {
	flag.Parse()
	start := time.Date(2021, 5, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2021, 6, 1, 0, 0, 0, 0, time.Local)
	params := map[string]interface{}{
		"start":    start.Format("2006-01-02"),
		"end":      end.Format("2006-01-02"),
		"uid":      uid,
		"region":   []string{"china"},
		"protocol": []string{"https", "http"},
		"type":     "bandwidth",
		"g":        "5min",
		"topn":     1,
	}

	var buf bytes.Buffer

	json.NewEncoder(&buf).Encode(params)

	resp, err := http.Post("http://deftonestraffic.fusion.internal.qiniu.io/v2/admin/traffic/user", "application/json", &buf)
	if err != nil {
		log.Fatal("http.Post:", err)
	}
	defer resp.Body.Close()

	var v struct {
		Total struct {
			Points []int64
		}
	}
	json.NewDecoder(resp.Body).Decode(&v)

	begin := start
	for _, p := range v.Total.Points {
		fmt.Printf("%s,%d\n", begin.Format("2006-01-02T15:04"), p)
		begin = begin.Add(5 * time.Minute)
	}
}
