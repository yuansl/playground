// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-05-31 08:23:20

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import "fmt"

type Generator interface {
	Next() bool
	Value() any
}

type chanGenerator struct {
	ch   chan any
	size int
}

func (g *chanGenerator) Next() bool { return g.size > 0 }

func (g *chanGenerator) Value() any {
	v, ok := <-g.ch
	if !ok {
		return nil
	}
	g.size--
	return v
}

func NewGenerator[S ~[]E, E any](items S) Generator {
	ch := make(chan any, len(items))
	for _, it := range items {
		ch <- it
	}
	close(ch)
	return &chanGenerator{size: len(items), ch: ch}
}

type Iterator interface {
	Filter(func(e any) bool) Iterator
	Map(func(e any) any) Iterator
	Reduce(state, first any) Iterator
	Min(less func(a, b any) bool) Iterator
	Collect() []any
}

type genIterator struct {
	generator Generator
}

func (it *genIterator) Filter(fn func(e any) bool) Iterator {
	items := []any{}

	for it.generator.Next() {
		v := it.generator.Value()

		if fn(v) {
			items = append(items, v)
		}
	}
	it.generator = NewGenerator(items)

	return it
}

func (it *genIterator) Map(fn func(e any) any) Iterator {
	items := []any{}

	for it.generator.Next() {
		v := it.generator.Value()
		items = append(items, fn(v))
	}
	it.generator = NewGenerator(items)

	return it
}

func (it *genIterator) Reduce(state any, first any) Iterator {
	panic("not implemented") // TODO: Implement
}

func (it *genIterator) Min(less func(a, b any) bool) Iterator {
	min := it.generator.Value()

	for it.generator.Next() {
		v := it.generator.Value()

		if less(v, min) {
			min = v
		}
	}
	it.generator = NewGenerator([]any{min})
	return it
}

func (it *genIterator) Collect() []any {
	items := []any{}

	for it.generator.Next() {
		items = append(items, it.generator.Value())
	}
	return items
}

func NewIterator(g Generator) Iterator {
	return &genIterator{generator: g}
}

func main() {
	gen := NewGenerator([]int{1, 2, 3, 3})

	lengths := NewIterator(gen).
		Filter(func(e any) bool {
			return len(e.(string)) > 1
		}).
		Min(func(a, b any) bool {
			return len(a.(string)) < len(b.(string))
		}).
		Map(func(e any) any {
			return len(e.(string))
		}).
		Collect()

	fmt.Println("lengths:", lengths)
}
