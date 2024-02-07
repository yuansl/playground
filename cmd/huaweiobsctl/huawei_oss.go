package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"

	"github.com/qbox/net-deftones/util"
	"github.com/yuansl/playground/oss"
)

const _HUAWEI_OBS_ENDPOINT_DEFAULT = "https://obs.cn-north-4.myhuaweicloud.com"

const _OBS_KEY_PREFIX = "qn-5min"

type huaweiOBS struct {
	*obs.ObsClient
}

// Download implements oss.ObjectStorageService.
func (hobs *huaweiOBS) Download(ctx context.Context, bucket string, key string) ([]byte, error) {
	result, err := hobs.GetObject(&obs.GetObjectInput{
		GetObjectMetadataInput: obs.GetObjectMetadataInput{
			Bucket: bucket, Key: key,
		}})
	if err != nil {
		return nil, fmt.Errorf("huawei.obs.GetObject: %w", err)
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

// List implements oss.ObjectStorageService.
func (hobs *huaweiOBS) List(ctx context.Context, bucket string, opts ...oss.ListOption) ([]oss.File, error) {
	var files []oss.File

	result, err := hobs.ListObjects(&obs.ListObjectsInput{
		Bucket:        bucket,
		ListObjsInput: obs.ListObjsInput{Prefix: _OBS_KEY_PREFIX},
	})
	if err != nil {
		return nil, fmt.Errorf("huawei.obs.ListObjects: %w", err)
	}
	for _, content := range result.Contents {
		files = append(files, oss.File{
			Owner:  content.Owner.ID,
			Bucket: bucket,
			Name:   filepath.Clean(content.Key),
			Size:   content.Size,
			Create: content.LastModified,
			Type:   oss.StorageTypeOf(string(content.StorageClass)),
		})
	}
	return files, nil
}

type ossUploadOptions struct {
	Key string
}

func WithKey(key string) oss.UploadOption {
	return oss.OptionFunc(func(opt any) {
		opt.(*ossUploadOptions).Key = key
	})
}

// Upload implements oss.ObjectStorageService.
func (hobs *huaweiOBS) Upload(ctx context.Context, bucket string, r io.Reader, opts ...oss.UploadOption) (*oss.UploadResult, error) {
	var options ossUploadOptions
	for _, opt := range opts {
		opt.Apply(&options)
	}
	fp, err := os.CreateTemp("/tmp/", "temp*")
	if err != nil {
		return nil, fmt.Errorf("os.CreateTemp: %w", err)
	}
	defer fp.Close()

	io.Copy(fp, r)

	input := obs.UploadFileInput{
		ObjectOperationInput: obs.ObjectOperationInput{Bucket: bucket, Key: options.Key},
		ContentType:          "text/plain",
		UploadFile:           fp.Name(),
	}
	result, err := hobs.UploadFile(&input)
	if err != nil {
		util.Fatal("huawei.obs.UploadFile:", err)
	}
	return &oss.UploadResult{Key: result.Key, Md5sum: result.ETag}, nil
}

func (hobs *huaweiOBS) Delete(ctx context.Context, bucket, key string) error {
	_, err := hobs.DeleteObject(&obs.DeleteObjectInput{Bucket: bucket, Key: key})
	if err != nil {
		return fmt.Errorf("huawei.obs.DeleteObject: %w", err)
	}
	return nil
}

var _ oss.ObjectStorageService = (*huaweiOBS)(nil)

type Option util.Option

type huaweiObsOptions struct {
	accessKey, secretKey string
	endpoint             string
}

func WithCredential(accessKey, secretKey string) Option {
	return util.OptionFunc(func(opt any) {
		opt.(*huaweiObsOptions).accessKey = accessKey
		opt.(*huaweiObsOptions).secretKey = secretKey
	})
}

func WithEndpoint(endpoint string) Option {
	return util.OptionFunc(func(opt any) {
		opt.(*huaweiObsOptions).endpoint = endpoint
	})
}

func NewHuaweiOSS(opts ...Option) oss.ObjectStorageService {
	var options huaweiObsOptions

	for _, opt := range opts {
		opt.Apply(&options)
	}
	if options.accessKey == "" || options.secretKey == "" {
		util.Fatal("NewOSS: Neither 'accessKey' nor 'secretKey' can be empty")
	}
	if options.endpoint == "" {
		options.endpoint = _HUAWEI_OBS_ENDPOINT_DEFAULT
	}
	obsClient, err := obs.New(options.accessKey, options.secretKey, options.endpoint)
	if err != nil {
		util.Fatal("obs.New error:", err)
	}
	return &huaweiOBS{ObsClient: obsClient}
}
