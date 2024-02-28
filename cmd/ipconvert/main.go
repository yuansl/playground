// -*- mode:go;mode:go-playground -*-
package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/qbox/net-deftones/logger"
	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/cdn"
	"golang.org/x/sync/errgroup"

	"github.com/yuansl/playground/oss/kodo"
	"github.com/yuansl/playground/util"
)

const (
	_ACCESS_KEY_DEFAULT = "557TpseUM8ovpfUhaw8gfa2DQ0104ZScM-BTIcBx"
	_SECRET_KEY_DEFAULT = "d9xLPyreEG59pR01sRQcFywhm4huL-XEpHHcVa90"
)

var _options struct {
	accessKey string // = _ACCESS_KEY_DEFAULT
	secretKey string // = _SECRET_KEY_DEFAULT
}

func init() {
	_options.accessKey = _ACCESS_KEY_DEFAULT
	_options.secretKey = _SECRET_KEY_DEFAULT
	if ak := os.Getenv("ACCESS_KEY"); ak != "" {
		_options.accessKey = ak
	}
	if sk := os.Getenv("SECRET_KEY"); sk != "" {
		_options.secretKey = sk
	}
}

type CloudStorage interface {
	Download(ctx context.Context, name string) ([]byte, error)
	Upload(ctx context.Context, filename string, expiry time.Duration) error
}

type CdnService interface {
	RefreshUrl(ctx context.Context, url string) error
}

func prepareAwdbs(ctx context.Context, store CloudStorage) error {
	wg, ctx := errgroup.WithContext(ctx)

	for _, key := range []string{"ipline.awdb", "ipv6.awdb", "ipv4.awdb", "ipv4.ipdb"} {
		wg.Go(func() error {
			logger.FromContext(ctx).Infof("Downloading file %s ...\n", key)
			data, err := store.Download(ctx, key)
			if err != nil {
				return fmt.Errorf("store.Download: %w", err)
			}
			err = os.WriteFile(key, data, 0600)
			if err != nil {
				return fmt.Errorf("os.WriteFile: %W", err)
			}
			return nil
		})
	}
	return wg.Wait()
}

func upload(ctx context.Context, files []string, oss CloudStorage) error {
	for _, file := range files {
		logger.FromContext(ctx).Infof("Uploding file %s ...\n", file)

		err := oss.Upload(ctx, file, 365*24*time.Hour)
		if err != nil {
			return fmt.Errorf("oss.Upload: %w", err)
		}
	}
	return nil
}

func ipconvert(ctx context.Context) error {
	executable := "/tmp/qip"
	err := prepareExecutable(executable)
	if err != nil {
		return fmt.Errorf("prepareExecutable: %w", err)
	}
	err = awdb2ipdb(ctx, executable)
	if err != nil {
		return fmt.Errorf("awdb2ipdb: %w", err)
	}
	return nil
}

func refreshCdnCache(ctx context.Context, filenames []string, cdn CdnService) error {
	for _, file := range filenames {
		url := "https://ipipfile.qbox.net/" + file

		logger.FromContext(ctx).Infof("Refreshing cdn cache of url %q ...\n", url)

		err := cdn.RefreshUrl(ctx, url)
		if err != nil {
			return fmt.Errorf("Cdn.RefreshUrl: %w", err)
		}
	}
	return nil
}

func main() {
	kodosrv := kodo.NewStorageService(
		kodo.WithLinkDomain("https://ipipfile.qbox.net"),
		kodo.WithCredential(_options.accessKey, _options.secretKey))
	oss := &kodoStorage{StorageService: kodosrv, bucket: "ipipfile"}
	cdnsrv := &fusion{cdn.NewCdnManager(auth.New(_options.accessKey, _options.secretKey))}
	outputs := []string{"neo.ipv4.ipdb", "neo.ipv6.ipdb"}
	ctx := logger.NewContext(context.Background(), logger.New())

	err := prepareAwdbs(ctx, oss)
	if err != nil {
		util.Fatal("prepareAwdbs:", err)
	}
	err = ipconvert(ctx)
	if err != nil {
		util.Fatal("ipconvert:", err)
	}
	err = upload(ctx, outputs, oss)
	if err != nil {
		util.Fatal("upload:", err)
	}
	err = refreshCdnCache(ctx, outputs, cdnsrv)
	if err != nil {
		util.Fatal("refreshCdnCache:", err)
	}
}
