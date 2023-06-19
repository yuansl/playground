package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

type Value struct {
	V int64 `json:"v"`
}

type Point struct {
	Time   time.Time `json:"time"`
	Values Value     `json:"values"`
}

var start = time.Date(2023, 5, 1, 0, 0, 0, 0, time.Local)

var points = [...]Point{
	{Time: start.Add(0), Values: Value{V: 9493439}},
	{Time: start.Add(5 * time.Minute), Values: Value{V: 343}},
	{Time: start.Add(10 * time.Minute), Values: Value{V: 924442}},
	{Time: start.Add(15 * time.Minute), Values: Value{V: 94932432}},
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		data, _ := httputil.DumpRequest(req, true)
		log.Printf("Received new request(raw): %q\n", data)

		w.Header().Set("Content-Type", "application/json; charset=utf8")
		json.NewEncoder(w).Encode(points)
	})
	println(http.ListenAndServe(":20051", nil))
}
