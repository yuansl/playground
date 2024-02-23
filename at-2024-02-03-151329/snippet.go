// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-02-03 15:13:29

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"

	"github.com/qbox/net-deftones/util"
)

type Pair struct {
	Key   string
	Value []byte
}

type Field Pair

type KeyOption util.Option

type Zmember struct {
	Score  float64
	Member []byte
}

type ZAttr struct {
	Change       bool
	Incr         bool
	AddNotExists bool
	AddExists    bool
}

type Set interface {
	Scard(key string)
	Sadd(key string, member []byte, members ...[]byte)
}

type ZrangeOptions struct {
	Withscores bool
	Limit      *struct {
		Offset int
		Count  int
	}
}

type ZrangeOption util.Option

func WithScores() ZrangeOption {
	return util.OptionFunc(func(opt any) { opt.(*ZrangeOptions).Withscores = true })
}

func WithLimit(offset, count int) ZrangeOption {
	return util.OptionFunc(func(opt any) {
		opt.(*ZrangeOptions).Limit = &struct{ Offset, Count int }{
			Offset: offset, Count: count,
		}
	})
}

type SortedSet interface {
	Zadd(key string, attr *ZAttr, member Zmember, members ...Zmember)
	Zrangebyscore(key string, score0, scoreN int, opts ...ZrangeOption)
	Zcard(key string) int
}

type Hash interface {
	HSet(key string, field string, value []byte, fields ...Field)
	HGet(key, field string)
}

type Redis interface {
	Get(key string) (any, error)
	Set(key string, value []byte, opts ...KeyOption) error
	Hash
	SortedSet
	Set
}

func main() {
	fmt.Println("Results:")
}
