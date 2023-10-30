package util

type Option interface {
	Apply(any)
}

// OptionFunc implements interface 'Option'
type OptionFunc func(any)

func (fn OptionFunc) Apply(o any) {
	fn(o)
}
