package main

//go:generate stringer -type Granularity -linecomment
type Granularity int

const (
	Granularity1hour Granularity = iota // hour
	Granularity5min                     // 5min
)

func GranularityOf(name string) Granularity {
	switch name {
	case "5min":
		return Granularity5min
	case "hour":
		return Granularity1hour
	default:
		panic("BUG: unknown granularity:" + name)
	}
}
