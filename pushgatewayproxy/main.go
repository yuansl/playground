package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

func main() {
	_url, err := url.Parse("http://localhost:9091")
	if err != nil {
		fatal("url.Parse:", err)
	}
	fatal(http.ListenAndServe(":9877", httputil.NewSingleHostReverseProxy(_url)))
}
