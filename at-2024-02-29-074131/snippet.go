// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-02-29 07:41:31

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"unsafe"

	"github.com/yuansl/playground/util"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <filename>\n", os.Args[0])
		os.Exit(1)
	}
	filename := os.Args[1]

	fp, err := os.Open(filename)
	if err != nil {
		util.Fatal("os.Open:", err)
	}
	defer fp.Close()

	var perCityWeather = make(map[string][]float64)

	for reader := bufio.NewReaderSize(fp, 32<<10); ; {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				util.Fatal("reader.ReadBytes:", err)
			}
			break
		}
		if bytes.HasPrefix(line, []byte{'#'}) {
			continue
		}
		line = bytes.TrimRight(line, "\n")
		s := *(*string)(unsafe.Pointer(&line))
		parts := strings.SplitN(s, ";", 2)
		temperature, _ := strconv.ParseFloat(parts[1], 32)
		perCityWeather[parts[0]] = append(perCityWeather[parts[0]], temperature)
	}
	type stat struct {
		station string
		Mean    float64
		Min     float64
		Max     float64
	}
	var statch = make(chan *stat)
	go func() {
		for city, weathers := range perCityWeather {
			sum := weathers[0]
			minimum := weathers[0]
			maximum := weathers[0]
			for i := 1; i < len(weathers); i++ {
				sum += weathers[i]
				minimum = min(minimum, weathers[i])
				maximum = max(maximum, weathers[i])
			}

			statch <- &stat{station: city, Mean: sum / float64(len(weathers)), Min: minimum, Max: maximum}
		}
		close(statch)
	}()
	var stats []*stat
	for stat := range statch {
		stats = append(stats, stat)

	}
	sort.Slice(stats, func(i, j int) bool { return stats[i].station < stats[j].station })
	fmt.Print("{")
	for i := range stats {
		fmt.Printf("%s=%.1f/%.1f/%.1f,", stats[i].station, stats[i].Min, stats[i].Mean, stats[i].Max)
	}
	fmt.Print("}\n")
}
