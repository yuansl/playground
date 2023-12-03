package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

func foo() {
	fmt.Printf("path: %q\n", filepath.Join("/usr/local", "../../tmp"))
	fmt.Printf("path: %q\n", filepath.Join(`D:\`, "Program files"))
}

func main() {
	for _, rurl := range []string{
		`https://www.example.com/path/to/file/你是谁/这是一个目录文件?e=1234567=`,
		`https://www.example.com/path/to/file/something?e=123`,
		`https://www.example.com/pictures/secret &and @home.avi?e=123&q=what%20is+the+best+practice%20of+golang`,
	} {
		_url, err := url.Parse(rurl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "url.Parse error: %v\n", err)
			continue
		}
		type alias url.URL
		fmt.Printf("raw-url=%q, _url: %+v, \n", rurl, (*alias)(_url))

		fmt.Printf("path: %q, escaped: %q, rawpath: %q\n", _url.Path, _url.EscapedPath(), _url.RawPath)

		const upperhex = "0123456789ABCDEF"

		fmt.Println("query:")
		for k, v := range _url.Query() {
			fmt.Printf("%q=%q\n", k, v)
		}

		__url, err := url.Parse(_url.EscapedPath())
		if err != nil {
			fmt.Printf("url.Parse error: %v\n", err)
		}
		fmt.Printf("__url.Path: %+v\n", __url.RawPath)
	}
}
