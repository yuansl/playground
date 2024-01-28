package oss

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/qbox/net-deftones/util"
)

//go:generate stringer -type StorageType -linecomment
type StorageType int

const (
	STORAGE_STANARD     StorageType = iota // standard
	STORAGE_LOWFREQ                        // low freq
	STORAGE_ARCHIVE                        // archive
	STORAGE_DEEPARCHIVE                    // deep archive
)

func StorageTypeOf(storage string) StorageType {
	for t := STORAGE_STANARD; t <= STORAGE_DEEPARCHIVE; t++ {
		if strings.EqualFold(t.String(), storage) {
			return t
		}
	}
	return -1
}

type File struct {
	Owner  string
	Bucket string
	Name   string
	Size   int64
	Md5sum string
	Create time.Time
	Type   StorageType
	Mime   string
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

type OptionFunc = util.OptionFunc

type ListOption util.Option

type ObjectStorageService interface {
	List(ctx context.Context, bucket string, opts ...ListOption) ([]File, error)
	Upload(ctx context.Context, bucket string, r io.Reader, opts ...UploadOption) (*UploadResult, error)
	Download(ctx context.Context, bucket, key string) ([]byte, error)
	Delete(ctx context.Context, bucket, key string) error
}
