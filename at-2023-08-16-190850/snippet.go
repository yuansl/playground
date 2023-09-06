// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-08-16 19:08:50

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
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
)

var blacklist_accesskeys = map[string]uint{
	"wD8gTca1WZzHn7P0fF-YRgoL_yI0HgbJ-OK7b4Dt": 1380485361,
}

func inBlacklist(ak string) (uint, bool) {
	for _ak, uid := range blacklist_accesskeys {
		if _ak == ak {
			return uid, true
		}
	}
	return 0, false
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		reverseProxy := &httputil.ReverseProxy{
			Rewrite: func(req *httputil.ProxyRequest) {
				data, err := httputil.DumpRequest(req.In, true)
				if err != nil {
					log.Printf("DumpRequest error: %v\n", err)
					w.WriteHeader(http.StatusBadRequest)
				}
				log.Printf("Received request(raw)=`%s`\n", data)

				req.Out.URL.Host = "localhost:20053"
				req.Out.URL.Scheme = "http"
				authorization := req.In.Header.Get("Authorization")
				if strings.HasPrefix(authorization, "Qiniu") {
					idx := strings.Index(authorization, "Qiniu")
					token := strings.TrimSpace(authorization[idx+5:])
					pieces := strings.Split(token, ":")
					if len(pieces) >= 2 {
						accesskey := pieces[0]
						if uid, yes := inBlacklist(accesskey); yes {
							fmt.Printf("access key=%q,uid=%d\n", accesskey, uid)
							req.Out.Header.Set("Authorization", "QiniuStub uid="+strconv.Itoa(int(uid))+"&ut=1")
						}
					}
				}
				req.SetXForwarded()

				fmt.Printf("Request header: '%v', Out header: '%v', out.URL: %+v, in.url.schema: %+v\n", req.In.URL, req.Out.Header, req.Out.URL.Host, req.In)

			},
		}
		reverseProxy.ServeHTTP(w, req)
	})
	go http.ListenAndServe(":20054", nil)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		data, err := httputil.DumpRequest(req, true)
		if err != nil {
			log.Printf("DumpRequest error: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		log.Printf("Received request(raw)=`%s`\n", data)

		w.Write([]byte(`hello, this is a server which listening on localhost:20053`))
	})
	err := http.ListenAndServe(":20053", mux)
	println(err)

	// err := http.ListenAndServe(":20053", &httputil.ReverseProxy{
	// 	Rewrite: func(req *httputil.ProxyRequest) {
	// 		authorization := req.In.Header.Get("Authorization")
	// 		if strings.HasPrefix(authorization, "Qiniu") {
	// 			idx := strings.Index(authorization, "Qiniu")
	// 			token := strings.TrimSpace(authorization[idx+5:])
	// 			pieces := strings.Split(token, ":")
	// 			if len(pieces) >= 2 {
	// 				accesskey := pieces[0]
	// 				if uid, yes := inBlacklist(accesskey); yes {
	// 					fmt.Printf("access key=%q,uid=%d\n", accesskey, uid)
	// 					req.Out.Header.Set("Authorization", "QiniuStub uid="+strconv.Itoa(int(uid))+"&ut=1")
	// 				}
	// 			}
	// 		}
	// 		fmt.Printf("Request header: '%v', Out header: '%v'\n", req.In.Header, req.Out.Header)
	// 	},
	// })
	// println(err)
}
