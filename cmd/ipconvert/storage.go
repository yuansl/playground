package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yuansl/playground/oss"
	"github.com/yuansl/playground/oss/kodo"
)

type kodoStorage struct {
	*kodo.StorageService
	bucket string
}

// Download implements Storage.
func (k *kodoStorage) Download(ctx context.Context, key string) ([]byte, error) {
	return k.StorageService.Download(ctx, k.bucket, key)
}

// Upload implements Storage.
func (k *kodoStorage) Upload(ctx context.Context, filename string, expiry time.Duration) error {
	key := filepath.Base(filename)
	fp, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("os.Open: %w", err)
	}
	defer fp.Close()

	options := []oss.UploadOption{kodo.WithKey(key)}
	if expiry > 0 {
		options = append(options, kodo.WithExpiry(expiry))
	}
	_, err = k.StorageService.Upload(ctx, k.bucket, bufio.NewReader(fp), options...)
	return err
}

var _ CloudStorage = (*kodoStorage)(nil)
