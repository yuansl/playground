package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
)

func foo() {
	fmt.Printf("path: %q\n", filepath.Join("/usr/local", "../../tmp"))
	fmt.Printf("path: %q\n", filepath.Join(`D:\`, "Program files"))
}

func main() {
	for _, _path := range []string{
		`https://www.example.com/path/to/file/你是谁/这是一个目录文件?e=1234567=`,
		`https://www.example.com/path/to/file/something?e=123`,
	} {
		_url, err := url.Parse(_path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "url.Parse error: %v\n", err)
		}
		type x url.URL
		fmt.Printf("_url: %+v\n", (*x)(_url))

		fmt.Printf("path: %q, escaped: %q, rawpath: %q\n", _url.Path, _url.EscapedPath(), _url.RawPath)

		const upperhex = "0123456789ABCDEF"

		c := '你'
		fmt.Printf("reflect.TypeOf(c): %v, c=`%#04x`\n", reflect.TypeOf(c), c)

		s := "你"
		for i := 0; i < len(s); i++ {
			fmt.Printf("%c in hex: %b%b\n", s[i], upperhex[s[i]>>4], upperhex[s[i]&15])
		}

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
