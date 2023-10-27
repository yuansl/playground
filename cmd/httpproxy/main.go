package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/yuansl/playground/util"
)

func main() {
	target, err := url.Parse("https://www.baidu.com/")
	if err != nil {
		util.Fatal("url.Parse:", err)
	}
	http.Handle("/", httputil.NewSingleHostReverseProxy(target))

	if err = http.ListenAndServe(":8080", nil); err != nil {
		util.Fatal(err)
	}
}
