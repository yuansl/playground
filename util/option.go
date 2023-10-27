package util

type Option interface {
	Apply(any)
}

// OptionFn implements interface 'Option'
type OptionFn func(any)

func (fn OptionFn) Apply(o any) {
	fn(o)
}
