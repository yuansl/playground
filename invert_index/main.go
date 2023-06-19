package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
)

type WordPair struct {
	Word       string
	DocumentID string
}

func main() {
	var resultq = make(chan WordPair, runtime.NumCPU())
	var perWordDocs = make(map[string][]string)

	go func() {
		defer close(resultq)

		var wg sync.WaitGroup
		for _, file := range os.Args[1:] {
			if file == "" {
				continue
			}

			wg.Add(1)
			go func(file string) {
				defer wg.Done()

				readWords(file, resultq)
			}(file)
			wg.Wait()
		}
	}()

	type groupKey struct {
		word string
		doc  string
	}

	var groupBy = make(map[groupKey]struct{})

	for pair := range resultq {
		key := groupKey{word: pair.Word, doc: pair.DocumentID}
		if _, exist := groupBy[key]; exist {
			continue
		}
		groupBy[key] = struct{}{}

		perWordDocs[pair.Word] = append(perWordDocs[pair.Word], pair.DocumentID)
	}
	for word, docs := range perWordDocs {
		fmt.Printf("%s -> %v\n", word, docs)
	}
}

func readWords(filename string, sendto chan<- WordPair) {
	fp, err := os.Open(filename)
	if err != nil {
		fatal("os.Open error: %v\n", err)
	}
	defer fp.Close()

	r := bufio.NewReader(fp)
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fatal("bufio.ReadBytes error: %v\n", err)
			}
			break
		}

		words := bytes.Fields(bytes.TrimSpace(line))
		for _, word := range words {
			sendto <- WordPair{
				Word:       string(word),
				DocumentID: filename,
			}
		}
	}
}

func fatal(format string, v ...any) {
	fmt.Fprintf(os.Stderr, format, v...)
	os.Exit(1)
}
