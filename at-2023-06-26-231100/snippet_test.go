package main

import "testing"

const N = 1000_000

func BenchmarkContinentAsRegions(b *testing.B) {
	b.Run("ContinentAsRegions", func(b *testing.B) {
		for i := 0; i < N; i++ {
			ContinentAsRegions["Africa"] = "X"
		}
	})
	b.Run("_ContinentAsRegions()", func(b *testing.B) {
		for i := 0; i < N; i++ {
			_ContinentAsRegions()["Africa"] = "X"
		}
	})
}
