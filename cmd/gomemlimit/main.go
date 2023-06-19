package main

import (
	"flag"
	"fmt"
	"net/http"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var allocobjsMetric = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "gomemlimit_allocated_objects",
	Help: "number of allocated objects of the program",
})

var (
	n    int
	addr string
)

func init() {
	flag.IntVar(&n, "n", 1000_000, "number of test")
	flag.StringVar(&addr, "addr", ":9090", "listen address of the service")
	flag.Parse()
}

func main() {
	var items = make(map[int]some)

	go exportMetrics()

	for i := 0; i < n; i++ {
		_some := some{
			Name: "say something, whatever you want",
			What: "what something",
			Some: "whatever",
		}
		if i%100 == 0 {
			delete(items, i%n)
		} else {
			items[i] = _some
		}
		allocobjsMetric.Add(+1)
	}

	fmt.Printf("allocated %d objects in total\n", n)

	syscall.Pause()
}

func exportMetrics() {
	prometheus.MustRegister(allocobjsMetric)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(addr, nil)
}

type some struct {
	Name string
	What string
	Some string
}
