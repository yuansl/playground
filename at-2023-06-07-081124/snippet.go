// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-06-07 08:11:24

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

type chanGenerator struct {
	ch  chan any
	it  int
	cap int
}

func (g *chanGenerator) Next() bool {
	return g.it < g.cap
}

func (g *chanGenerator) Value() any {
	v, ok := <-g.ch
	if !ok {
		return nil
	}
	g.it++
	return v
}

func NewGenerator(items []any) Generator {
	ch := make(chan any, len(items))
	for _, it := range items {
		ch <- it
	}
	close(ch)

	return &chanGenerator{ch: ch, cap: len(items), it: 0}
}

type iterator struct {
	g Generator
}

func (it *iterator) Max(less func(first, second any) bool) any {
	max := it.g.Value()

	for it.g.Next() {
		value := it.g.Value()
		if less(max, value) {
			max = value
		}
	}
	return max
}
func (it *iterator) Min(less func(first, second any) bool) any {
	min := it.g.Value()

	for it.g.Next() {
		value := it.g.Value()
		if less(value, min) {
			min = value
		}
	}
	return min
}

func (it *iterator) Map(fn func(v any) any) {
	values := []any{}

	for it.g.Next() {
		value := it.g.Value()
		values = append(values, fn(value))
	}
	it.g = NewGenerator(values)
}

func NewIterator(g Generator) *iterator {
	return &iterator{g: g}
}

func main() {
	items := []any{99, -1, -183, 3, 3, 43, 03, 343}

	g := generator.NewSlice(items)
	// g := NewGenerator()

	// it := NewIterator(g)

	// max := it.Min(func(first, second any) bool {
	// 	return first.(int) < second.(int)
	// })

	it := goterator.New(g)

	it = it.Map(func(e any) any {
		return e.(int) * 11
	})
	it.ForEach(goterator.IterFunc(func(e any) {
		fmt.Println("element:", e)
	}))
}
