package main

import "strings"

func main() {

}

var suffixes = []string{
	"qiniu.com",
	"qiniu.io",
	"qiniuapi.com",
	"qbox.me",
}

func isQiniuDomain(domain string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(domain, suffix) {
			return true
		}
	}
	return false
}
