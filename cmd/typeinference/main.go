package main

import (
	"fmt"
	"net"
	"net/netip"
	"os"
	"reflect"
	"regexp"
)

type Numeric interface {
	~int | ~int32 | ~int64
}

func sum[E Numeric](slice []E) E {
	var s E
	for _, v := range slice {
		s += v
	}
	return s
}

type Point []int64

func fatal(format string, v ...any) {
	fmt.Fprintf(os.Stderr, "fatal error: "+format, v...)
	os.Exit(1)
}

func foo_netip() {
	xs := sum([]int{1, 2, 3, 4})
	y := sum(Point([]int64{3, 4, 5, 6, int64(xs)}))
	sum([]int32{1, 2, 2, 3, int32(y)})

	cases := []struct {
		network  string
		address  string
		resolver func(network, address string) (net.Addr, error)
	}{
		{
			network: "ip",
			address: "www.baidu.com",
			resolver: func(network, address string) (net.Addr, error) {
				return net.ResolveIPAddr(network, address)
			},
		},
		{
			network: "tcp",
			address: "www.baidu.com:443",
			resolver: func(network, address string) (net.Addr, error) {
				return net.ResolveTCPAddr(network, address)
			},
		},
	}
	for _, c := range cases {
		addr, err := c.resolver(c.network, c.address)
		if err != nil {
			fatal("net.ResolveIPAddr error: %v\n", err)
		}
		fmt.Printf("resolved address %q: %+v\n", c.address, addr)
	}

	addr, err := netip.ParseAddr("127.0.0.1")
	if err != nil {
		fatal("netip.ParseAddr error: %v\n", err)
	}
	fmt.Printf("addr: %+v\n", addr)

	ns, err := net.LookupNS("www.baidu.com")
	if err != nil {
		fatal("net.LookupNS error: %v\n", err)
	}
	for _, n := range ns {
		fmt.Printf("NS %+v\n", n)
	}
}

type Some struct{}

func (Some) Print1()  {}
func (*Some) Print2() {}

type Printer interface {
	Print1()
}

type Printer2 interface {
	Print2()
}

func implements(v any, _type reflect.Type) bool {
	return reflect.ValueOf(v).Type().Implements(_type)
}

func TestImplements() {
	for _, v := range []any{
		struct{ Some }{},
		&struct{ Some }{},
		struct{ *Some }{},
		&struct{ *Some }{},
	} {
		i1 := implements(v, reflect.TypeOf((*Printer)(nil)).Elem())
		i2 := implements(v, reflect.TypeOf((*Printer2)(nil)).Elem())

		fmt.Printf("%#v implements Printer type: %t, s implements Printer2 type: %t\n",
			v, i1, i2)
	}
}

func main() {
	matched, err := regexp.MatchString(`(?:.*\bword1\b)\s+(?:.*word2)(?:.*\bword3\b)`, "word1 word2 word3")
	if err != nil {
		fatal("regexp.MatchString: %v", err)
	}
	fmt.Printf("matched: %t\n", matched)
}
