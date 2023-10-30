package oss

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/yuansl/playground/util"
)

type File struct {
	Name   string
	Size   int64
	Md5sum string
}

type Stat struct {
	Name   int64
	Size   int64
	Mod    time.Time
	Access time.Time
	Md5sum string
}

type UploadResult struct {
	Url    string
	Key    string
	Size   int
	Md5sum string
}

func (r *UploadResult) String() string {
	return fmt.Sprintf("\n  key: %q\n  size: %d\n  md5sum: %s\n  url: %q\n", r.Key, r.Size, r.Md5sum, r.Url)
}

type UploadOption util.Option

type ListOption util.Option

type ObjectStorageService interface {
	List(ctx context.Context, opts ...ListOption) ([]File, error)
	Stat(ctx context.Context, name string)
	Upload(ctx context.Context, r io.Reader, opts ...UploadOption) (*UploadResult, error)
	Download(ctx context.Context, key string) ([]byte, error)
}
