package main

import (
	"testing"
)

func FuzzSome(f *testing.F) {
	for _, domain := range []string{"www.qiniu.com", "www.qiniuapi.com", "defy-dsync.qiniuapi.com", "pili.qiniuapi.com"} {
		f.Add(domain)
	}
	f.Fuzz(func(t *testing.T, domain string) {
		t.Logf("input domain `%s`\n", domain)
		if !isQiniuDomain(domain) {
			t.Logf("domain `%s` is not a qiniu domain\n", domain)
		}
	})
}

func FuzzHello(f *testing.F) { f.Skip() }
