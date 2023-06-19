// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-05-29 22:50:44

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"

	"github.com/yaa110/goterator"
	"github.com/yaa110/goterator/generator"
)

type Generator interface {
	Next() bool
	Value() any
}

type xgenerator struct {
	gen    chan any
	size   int
	closed bool
}

func (g *xgenerator) Next() bool {
	return !g.closed && g.size > 0
}

func (g *xgenerator) Value() any {
	v, ok := <-g.gen
	if !ok {
		g.closed = true
		return nil
	}
	g.size--

	return v
}

func NewSlice(feeds []any) Generator {
	gen := xgenerator{gen: make(chan any, len(feeds)), size: len(feeds), closed: false}

	for _, feed := range feeds {
		gen.gen <- feed
	}
	close(gen.gen)
	return &gen
}

func mockGeneroator() {
	words := []any{"an", "example", "of", "goterator"}
	gen := NewSlice(words)

	for gen.Next() {
		fmt.Println("next:", gen.Value())
	}
}

type Iterator interface {
	Map(func(any) any) Iterator
	Collection() []any
}

type iterator struct {
	ingen  Generator
	outgen Generator
}

func (it *iterator) Map(f func(any) any) Iterator {
	values := []any{}
	for it.ingen.Next() {
		v := f(it.ingen.Value())
		values = append(values, v)
	}
	it.outgen = NewSlice(values)
	return it
}

func (it *iterator) Collection() []any {
	values := []any{}

	for it.outgen.Next() {
		values = append(values, it.outgen.Value())
	}
	return values
}

func NewIterator(ingen Generator) Iterator {
	return &iterator{ingen: ingen}
}

func main() {
	words := []interface{}{"an", "example", "of", "goterator"}
	gen := generator.NewSlice(words)
	lengths := goterator.New(gen).Map(func(word interface{}) interface{} {
		return len(word.(string))
	}).Collect()
	// lengths := NewIterator(NewSlice(words)).Map(func(word interface{}) interface{} {
	// 	return len(word.(string))
	// }).Collection()

	fmt.Println("lengths:", lengths)
}
