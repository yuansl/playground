package domaininfo

import "testing"

func BenchmarkTransform(b *testing.B) {
	b.Run("transform", func(b *testing.B) {
		var domain = DomainInfo{
			Name:     "www.example.com.abcdefghijklmn.fusion-conflict",
			Conflict: true,
		}
		transform(&domain)
	})
	b.Run("nameWithoutConflict", func(b *testing.B) {
		var domain = DomainInfo{
			Name:     "www.example.com.abcdefghijklmn.fusion-conflict",
			Conflict: true,
		}
		nameWithoutConflict(&domain)
	})
}
