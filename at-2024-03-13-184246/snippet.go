// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-03-13 18:42:46

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/yuansl/playground/util"
)

type Dimensions struct {
	Timestamp     time.Time `json:"datetimeFiveMinutes"`
	Host          string    `json:"clientRequestHTTPHost"`
	ClientCountry string    `json:"clientCountryName"`

	// for cloudflare cache status
	// see https://developers.cloudflare.com/cache/concepts/cache-responses/
	CacheStatus CacheStatus `json:",string"` // oneof: hit|miss|none|unknown|dynamic...
}

//go:generate stringer -type CacheStatus -linecomment
type CacheStatus int

const (
	CacheHit     CacheStatus = iota // hit
	CacheDynmaic                    // dynamic
	CacheMiss                       // miss
	CacheNone                       // none
)

func main() {
	var dim Dimensions

	err := json.Unmarshal([]byte(`{"datetimeFiveMinutes":"2024-03-01T00:00:00Z", "clientRequestHTTPHost":"www.example.com", "clientCountryName":"CN", "cacheStatus":"2"}`), &dim)
	if err != nil {
		util.Fatal(err)
	}
	fmt.Printf("Results: %+v\n", dim)
}
