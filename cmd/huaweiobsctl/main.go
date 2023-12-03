// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-11-21 15:00:53

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yuansl/playground/logger"
	"github.com/yuansl/playground/oss"
	"github.com/yuansl/playground/util"
)

const _DOWNLOAD_WATER_MARK_SIZE = 0

var (
	_accessKey string
	_secretKey string
)

// 存储桶名称
const _BUCKET_NAME = "cdn-3rd-qn-bj4"

func init() {
	_accessKey = os.Getenv("ACCESS_KEY")
	_secretKey = os.Getenv("SECRET_KEY")
}

func upload(ctx context.Context, _oss oss.ObjectStorageService) {
	fp, err := os.CreateTemp("/tmp", "huawei-obs-test*")
	if err != nil {
		util.Fatal("os.CreateTemp: '%v'\n", err)
	}
	defer fp.Close()

	fmt.Fprintf(fp, "this is a test message")

	result, err := _oss.Upload(ctx, _BUCKET_NAME, fp, WithKey("qn-5min/test"))
	if err != nil {
		util.Fatal("obsClient.UploadFile:", err)
	}
	fmt.Printf("UploadFile result: %+v\n", result)
}

func main() {
	huaweiOSS := NewHuaweiOSS(WithCredential(_accessKey, _secretKey))
	ctx := logger.NewContext(context.TODO(), logger.New())

	files, err := huaweiOSS.List(ctx, _BUCKET_NAME)
	if err != nil {
		util.Fatal("oss.List:", err)
	}
	for i, file := range files {
		fmt.Printf("	#%d: file: '%+v'\n", i, file)

		if file.Size >= _DOWNLOAD_WATER_MARK_SIZE && file.Name != _OBS_KEY_PREFIX {
			data, err := huaweiOSS.Download(ctx, _BUCKET_NAME, file.Name)
			if err != nil {
				util.Fatal("oss.Download(bucket='%s',key='%s') error: %v\n", _BUCKET_NAME, file.Name, err)
			}

			saveas, err := os.OpenFile(filepath.Join("/tmp", filepath.Base(file.Name)), os.O_CREATE|os.O_APPEND|os.O_WRONLY|os.O_TRUNC, 0600)
			if err != nil {
				util.Fatal("os.OpenFile:", err)
			}
			defer saveas.Close()

			saveas.Write(data)

			if file.Name == "qn-5min/test" {
				fmt.Printf("		The Key '%s' in bucket '%s' will be deleted ...\n", file.Name, file.Bucket)

				if err := huaweiOSS.Delete(ctx, file.Bucket, file.Name); err != nil {
					util.Fatal("oss.Delete:", err)
				}
			}
		}
	}
}
