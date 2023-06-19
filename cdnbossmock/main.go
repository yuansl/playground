package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
)

var addr string

func init() {
	flag.StringVar(&addr, "addr", ":8082", "cdnboss mock interface")

	http.HandleFunc("/v2/metric", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		data, _ := httputil.DumpRequest(req, true)
		log.Printf("request(raw): `%s`\n", string(data))

		w.Write([]byte(`{"status": "ok"}`))
	})
}

func main() {
	flag.Parse()

	log.Println(http.ListenAndServe(addr, nil))
}
