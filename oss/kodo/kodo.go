package kodo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/storage"

	"github.com/yuansl/playground/oss"
	"github.com/yuansl/playground/util"
)

const (
	_KODO_SERVICE_ENDPOINT_DEFAULT   = "http://pybwef48y.bkt.clouddn.com"
	_KODO_RESPONSE_BODY_SIZE_MAX     = 1 << 30   // 1GiB
	_KODO_RANGE_REQUEST_TRIGGER_SIZE = 128 << 20 // 128MiB
	_KODO_BLOCK_SIZE                 = 4 << 20   // 4MiB

	_KODO_FILE_EXPIRY_DEFAULT = 60 * time.Minute
)

// storageService implements ObjectStorageService.
type storageService struct {
	credentials   *auth.Credentials
	uploader      *storage.ResumeUploader
	bucketManager *storage.BucketManager
	bucket        string
	endpoint      string
}

var _ oss.ObjectStorageService = (*storageService)(nil)

type UrlOption util.Option

type privateUrlOptions struct {
	expiry time.Duration
}

func WithPrivateUrlExpiry(expiry time.Duration) UrlOption {
	return util.OptionFunc(func(op any) {
		op.(*privateUrlOptions).expiry = expiry
	})
}

func (kodo *storageService) UrlOfKey(ctx context.Context, key string, opts ...UrlOption) string {
	var options privateUrlOptions
	for _, op := range opts {
		op.Apply(&options)
	}
	if options.expiry <= 0 {
		options.expiry = _KODO_FILE_EXPIRY_DEFAULT
	}
	return storage.MakePrivateURL(kodo.credentials, kodo.endpoint, key, time.Now().Add(options.expiry).Unix())
}

type DownloadOption util.Option

type downloadOptions struct{ rangeBegin, rangeEnd int }

func WithRangeRequest(begin, end int) DownloadOption {
	return util.OptionFunc(func(opt any) {
		opt.(*downloadOptions).rangeBegin = begin
		opt.(*downloadOptions).rangeEnd = end
	})
}

func doOnceHttpRequest(ctx context.Context, url string, opts ...DownloadOption) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %v", err)
	}
	var options downloadOptions
	for _, op := range opts {
		op.Apply(&options)
	}
	if options.rangeBegin >= 0 && options.rangeEnd > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", options.rangeBegin, options.rangeEnd))
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http.Client.Do: %v", err)
	}
	return io.ReadAll(io.LimitReader(res.Body, _KODO_RESPONSE_BODY_SIZE_MAX))
}

func (kodo *storageService) Download(ctx context.Context, key string) ([]byte, error) {
	url := kodo.UrlOfKey(ctx, key)
	res, err := http.Head(url)
	if err != nil {
		return nil, fmt.Errorf("http.Head: %v", err)
	}
	if !strings.Contains(res.Header.Get("Accept-Ranges"), "bytes") || res.ContentLength < _KODO_RANGE_REQUEST_TRIGGER_SIZE {
		return doOnceHttpRequest(ctx, url)
	}

	buf := make([]byte, res.ContentLength)
	errorq := make(chan error, 1)

	go func() {
		defer close(errorq)
		var wg sync.WaitGroup
		var climit = make(chan struct{}, runtime.NumCPU())

		for off := 0; off <= int(res.ContentLength); off += _KODO_BLOCK_SIZE {
			begin, end := off, off+_KODO_BLOCK_SIZE
			if end > int(res.ContentLength) {
				end = int(res.ContentLength)
			}

			climit <- struct{}{}
			wg.Add(1)
			go func(begin, end int) {
				defer func() {
					<-climit
					wg.Done()
				}()

				data, err := doOnceHttpRequest(ctx, url, WithRangeRequest(begin, end))
				if err != nil {
					errorq <- err
					return
				}
				if len(data) != end-begin+1 {
					errorq <- fmt.Errorf("mismatch: file: %s, got %d bytes data, but expected %d (Range: bytes=%d-%d)\n", key, len(data), end-begin+1, begin, end)
					return
				}

				copy(buf[begin:end+1], data)

			}(begin, end-1)
		}
		wg.Wait()
	}()
	for {
		select {
		case <-ctx.Done():
			return buf, ctx.Err()
		case err := <-errorq:
			if err != nil {
				return buf, fmt.Errorf("doOnceHttpRequest: %v", err)
			}
			return buf, nil
		}
	}
}

type listOptions struct {
	limit  int
	prefix string
}

func WithListLimit(limit int) oss.ListOption {
	return util.OptionFunc(func(opt any) {
		opt.(*listOptions).limit = limit
	})
}

func WithListPrefix(prefix string) oss.ListOption {
	return util.OptionFunc(func(opt any) {
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
	if options.limit <= 0 {
		options.limit = 1
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
	return util.OptionFunc(func(op any) {
		op.(*uploadOptions).key = key
	})
}

func WithSaveBucket(bucket string) oss.UploadOption {
	return util.OptionFunc(func(op any) {
		op.(*uploadOptions).bucket = bucket
	})
}

func WithExpiry(expiry time.Duration) oss.UploadOption {
	return util.OptionFunc(func(op any) {
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
	return util.OptionFunc(func(op any) {
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

type ServiceOption util.Option

type serviceOptions struct {
	accessKey string
	secretKey string
	endpoint  string
	bucket    string
}

func WithCredential(accessKey, secretKey string) ServiceOption {
	return util.OptionFunc(func(opt any) {
		opt.(*serviceOptions).accessKey = accessKey
		opt.(*serviceOptions).secretKey = secretKey
	})
}

func WithEndpoint(endpoint string) ServiceOption {
	return util.OptionFunc(func(opt any) {
		opt.(*serviceOptions).endpoint = endpoint
	})
}

func WithBucket(bucket string) ServiceOption {
	return util.OptionFunc(func(opt any) { opt.(*serviceOptions).bucket = bucket })
}

func NewStorageService(opts ...ServiceOption) *storageService {
	var options serviceOptions

	for _, op := range opts {
		op.Apply(&options)
	}
	if options.endpoint == "" {
		options.endpoint = _KODO_SERVICE_ENDPOINT_DEFAULT
	}
	credentials := auth.New(options.accessKey, options.secretKey)
	return &storageService{
		credentials:   credentials,
		bucket:        options.bucket,
		bucketManager: storage.NewBucketManager(credentials, nil),
		uploader:      storage.NewResumeUploader(&storage.Config{}),
		endpoint:      options.endpoint,
	}
}
