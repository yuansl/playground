package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
)

func download(ctx context.Context, url string, saveas io.Writer) error {
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("%w: http.Get(url=%q): %v", ErrProtocol, url, err)
	}
	defer res.Body.Close()

	gz, err := gzip.NewReader(res.Body)
	if err != nil {
		return fmt.Errorf("%w: gzip.NewReader(res.Body): %v", ErrProtocol, err)
	}
	_, err = io.Copy(saveas, gz)
	return err
}
