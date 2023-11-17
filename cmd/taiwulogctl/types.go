package main

//go:generate stringer -type Region -linecomment
type Region int

const (
	RegionChina Region = iota + 1 // china
)

//go:generate stringer -type DataType -linecomment
type DataType int

const (
	DataTypeBandwidth DataType = iota // bandwidth
)
