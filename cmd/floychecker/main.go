package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"text/template"
)

var (
	filename  string
	templatef string
)

func init() {
	flag.StringVar(&filename, "f", "", "filename in csv format")
	flag.StringVar(&templatef, "template", "", "template file name")
}

func main() {
	flag.Parse()

	fp, err := os.Open(filename)
	if err != nil {
		fmt.Fprint(os.Stderr, "os.Open:", err)
		os.Exit(1)
	}

	loadVarParams(fp)
}

// csv format see https://tools.ietf.org/html/rfc4180
type dictCsvReader struct {
	csvr    *csv.Reader
	headers []string
}

func NewDictCsvReader(r io.Reader) (*dictCsvReader, error) {
	csvr := csv.NewReader(r)

	headers, err := csvr.Read()
	if err != nil {
		return nil, err
	}

	return &dictCsvReader{
		csvr:    csvr,
		headers: headers,
	}, nil
}

type DictCsvField map[string]string

func (r *dictCsvReader) Read() (DictCsvField, error) {
	fields := make(map[string]string)

	records, err := r.csvr.Read()
	if err != nil {
		return nil, err
	}

	for i, record := range records {
		fields[r.headers[i]] = record
	}

	return fields, nil
}

type FloyConfig struct {
	Node string
	Env  string
	Pkg  string
}

func loadVarParams(r io.Reader) {
	reader, err := NewDictCsvReader(r)
	if err != nil {
		fmt.Fprint(os.Stderr, "NewDictCsvReader failed:", err)
		os.Exit(2)
	}

	t := template.Must(template.New("floy").Parse("deploy package `{{.Pkg}}` of `env_{{.Env}}` on node {{.Node}}\n"))

	for {
		records, err := reader.Read()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fmt.Println("csv.Read failed: ", err)
			}
			break
		}

		if err := t.Execute(os.Stdout, FloyConfig{
			Node: records["node"],
			Env:  records["env"],
			Pkg:  records["pkg"],
		}); err != nil {
			fmt.Printf("template.Execute failed: %v\n", err)
		}
	}

}
