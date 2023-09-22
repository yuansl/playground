package kodo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
	"unsafe"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/storage"

	"github.com/yuansl/playground/cmd/kodoctl/oss"
	"github.com/yuansl/playground/utils"
)

const (
	_KODO_STORAGE_ENDPOINT = "http://pybwef48y.bkt.clouddn.com"
	_DOWNLOAD_SIZE_MAX     = 1 << 30
	_EXPIRY_DEFAULT        = 60 * time.Minute
)

// storageService implements ObjectStorageService.
type storageService struct {
	credentials   *auth.Credentials
	uploader      *storage.ResumeUploader
	bucketManager *storage.BucketManager
	bucket        string
}

var _ oss.ObjectStorageService = (*storageService)(nil)

type UrlOption utils.Option

type privateUrlOptions struct {
	expiry time.Duration
}

func WithPrivateUrlExpiry(expiry time.Duration) UrlOption {
	return utils.OptionFn(func(op any) {
		op.(*privateUrlOptions).expiry = expiry
	})
}

func (kodo *storageService) UrlOfKey(ctx context.Context, key string, opts ...UrlOption) string {
	var options privateUrlOptions
	for _, op := range opts {
		op.Apply(&options)
	}
	if options.expiry <= 0 {
		options.expiry = _EXPIRY_DEFAULT
	}
	return storage.MakePrivateURL(kodo.credentials, _KODO_STORAGE_ENDPOINT, key, time.Now().Add(options.expiry).Unix())
}

func (kodo *storageService) Download(ctx context.Context, key string) ([]byte, error) {
	url := kodo.UrlOfKey(ctx, key)
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http.Get: %v", err)
	}
	return io.ReadAll(io.LimitReader(res.Body, _DOWNLOAD_SIZE_MAX))
}

type listOptions struct {
	limit  int
	prefix string
}

func WithListLimit(limit int) oss.ListOption {
	return utils.OptionFn(func(opt any) {
		opt.(*listOptions).limit = limit
	})
}

func WithListPrefix(prefix string) oss.ListOption {
	return utils.OptionFn(func(opt any) {
		opt.(*listOptions).prefix = prefix
	})
}

// List implements ObjectStorageService.
func (kodo *storageService) List(ctx context.Context, opts ...oss.ListOption) ([]oss.File, error) {
	var files []oss.File
	var options listOptions

	for _, op := range opts {
		op.Apply(&options)
	}
	inputOptions := []storage.ListInputOption{storage.ListInputOptionsLimit(1000)}
	if options.prefix != "" {
		inputOptions = append(inputOptions, storage.ListInputOptionsPrefix(options.prefix))
	}
	for marker := ""; ; {
		res, hasNext, err := kodo.bucketManager.ListFilesWithContext(ctx, kodo.bucket,
			append(inputOptions, storage.ListInputOptionsMarker(marker))...,
		)
		if err != nil {
			return nil, fmt.Errorf("kodo.bucket.List(bucket='%s'): %v", kodo.bucket, err)
		}
		for _, it := range res.Items {
			files = append(files, oss.File{Name: it.Key, Size: it.Fsize, Md5sum: it.Md5})
			if len(files) >= options.limit {
				goto endit
			}
		}
		if !hasNext || res.Marker == "" {
			break
		}
		marker = res.Marker
	}
endit:
	return files, nil
}

// Stat implements ObjectStorageService.
func (*storageService) Stat(ctx context.Context, name string) {
	panic("unimplemented")
}

type NotifyFn func(blkIdx int, blkSize int, res *storage.BlkputRet)

type uploadOptions struct {
	key    string
	bucket string
	expiry time.Duration
	notify NotifyFn
}

func WithKey(key string) oss.UploadOption {
	return utils.OptionFn(func(op any) {
		op.(*uploadOptions).key = key
	})
}

func WithSaveBucket(bucket string) oss.UploadOption {
	return utils.OptionFn(func(op any) {
		op.(*uploadOptions).bucket = bucket
	})
}

func WithExpiry(expiry time.Duration) oss.UploadOption {
	return utils.OptionFn(func(op any) {
		op.(*uploadOptions).expiry = expiry
	})
}

// BlockputResult 表示分片上传每个片上传完毕的返回值
type BlockputResult struct {
	Ctx        string `json:"ctx"`
	Checksum   string `json:"checksum"`
	Crc32      uint32 `json:"crc32"`
	Offset     uint32 `json:"offset"`
	Host       string `json:"host"`
	ExpiredAt  int64  `json:"expired_at"`
	chunkSize  int
	fileOffset int64
	blkIdx     int
}

func WithNotify(notify func(blkId, blkSize, offset int)) oss.UploadOption {
	return utils.OptionFn(func(op any) {
		op.(*uploadOptions).notify = func(blkIdx, blkSize int, res *storage.BlkputRet) {
			r := (*BlockputResult)(unsafe.Pointer(res))
			notify(blkIdx, blkSize, int(r.fileOffset))
		}
	})
}

func (kodo *storageService) GetUploadToken(bucket string, expiry time.Duration) string {
	policy := storage.PutPolicy{Scope: bucket}

	if expiry > 0 {
		policy.Expires = uint64(time.Now().Add(expiry).Unix())
	}

	return policy.UploadToken(kodo.credentials)
}

// Upload implements ObjectStorageService.
func (kodo *storageService) Upload(ctx context.Context, reader io.Reader, opts ...oss.UploadOption) (*oss.UploadResult, error) {
	var options uploadOptions

	for _, op := range opts {
		op.Apply(&options)
	}
	if options.bucket == "" {
		options.bucket = kodo.bucket
	}
	token := kodo.GetUploadToken(options.bucket, options.expiry)

	md5reader := NewMd5InterceptReader(reader)

	if err := kodo.uploader.PutWithoutSize(ctx, nil, token, options.key, md5reader, &storage.RputExtra{Notify: options.notify}); err != nil {
		return nil, fmt.Errorf("kodo.ResumeUploader.PutWithoutSize: %v", err)
	}

	st, err := kodo.bucketManager.Stat(options.bucket, options.key)
	if err != nil {
		return nil, fmt.Errorf("kodo.BucketManager.Stat: %v", err)
	}
	return &oss.UploadResult{
		Key:    options.key,
		Size:   int(st.Fsize),
		Md5sum: md5reader.Sum(),
		Url:    kodo.UrlOfKey(ctx, options.key, WithPrivateUrlExpiry(options.expiry)),
	}, nil
}

func NewStorageService(accessKey, secretKey, bucket string) *storageService {
	credentials := auth.New(accessKey, secretKey)
	return &storageService{
		credentials:   credentials,
		bucket:        bucket,
		bucketManager: storage.NewBucketManager(credentials, nil),
		uploader:      storage.NewResumeUploader(&storage.Config{}),
	}
}
