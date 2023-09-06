package main

type Option interface {
	Apply(any)
}

type OptionFunc func(any)

func (of OptionFunc) Apply(op any) {
	of(op)
}
