package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/yuansl/playground/cmd/kodoctl/oss"
	"github.com/yuansl/playground/cmd/kodoctl/oss/kodo"
	"github.com/yuansl/playground/util"
	"github.com/yuansl/playground/utils"
)

const (
	_ACCESS_KEY_DEFAULT           = "557TpseUM8ovpfUhaw8gfa2DQ0104ZScM-BTIcBx"
	_SECRET_KEY_DEFAULT           = "d9xLPyreEG59pR01sRQcFywhm4huL-XEpHHcVa90"
	_KODO_BUCKET_DEFAULT          = "fusionlogtest"
	_KODO_FILE_KEY_PREFIX_DEFAULT = "share"
)

var (
	_accessKey = _ACCESS_KEY_DEFAULT
	_secretKey = _SECRET_KEY_DEFAULT
	_bucket    string
	_prefix    string
	_limit     int
	_filename  string
	_key       string
	_expiry    time.Duration
	_output    string
)

func init() {
	if ak := os.Getenv("ACCESS_KEY"); ak != "" {
		_accessKey = ak
	}
	if sk := os.Getenv("SECRET_KEY"); sk != "" {
		_secretKey = sk
	}
}

func parseCmdArgs(args []string) {
	flags := flag.NewFlagSet("kodoctl", flag.ExitOnError)

	flags.StringVar(&_bucket, "bucket", _KODO_BUCKET_DEFAULT, "specify the kodo bucket")
	flags.StringVar(&_prefix, "prefix", _KODO_FILE_KEY_PREFIX_DEFAULT, "specify prefix of a object file")
	flags.IntVar(&_limit, "limit", 5, "list <limit> files at most")
	flags.StringVar(&_filename, "file", "", "specify a local file for uploading")
	flags.StringVar(&_key, "key", "", "specify file key in oss/kodo")
	flags.DurationVar(&_expiry, "expiry", 24*time.Hour, "specify expiry date of a file storead in oss")
	flags.StringVar(&_output, "o", "", "save as ...")
	if err := flags.Parse(args[1:]); err != nil {
		utils.Fatal(err)
	}
}

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("Usage: %s list|download|upload\n", os.Args[0])
		os.Exit(0)
	}
	parseCmdArgs(os.Args[1:])

	storage := kodo.NewStorageService(_accessKey, _secretKey, _bucket)

	ctx := context.TODO()

	switch action := os.Args[1]; action {
	case "download":
		if _key == "" {
			utils.Fatal("kodo: oss key must not be empty")
		}
		data, err := storage.Download(ctx, _key)
		if err != nil {
			utils.Fatal("store.Download error:", err)
		}
		output := os.Stdout
		if _output != "" {
			output, err = os.OpenFile(_output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				utils.Fatal("os.OpenFile:", err)
			}
			defer output.Close()
		}
		if _, err = output.Write(data); err != nil {
			utils.Fatal("os.Write:", err)
		}

	case "list":
		var options []oss.ListOption
		if _prefix != "" {
			options = append(options, kodo.WithListPrefix(_prefix))
		}
		if _limit > 0 {
			options = append(options, kodo.WithListLimit(_limit))
		}
		files, err := storage.List(ctx, options...)
		if err != nil {
			utils.Fatal(err)
		}
		for _, it := range files {
			fmt.Printf("file: %+v\n", it)
		}
	case "upload":
		fp, err := os.Open(_filename)
		if err != nil {
			util.Fatal(err)
		}
		defer fp.Close()

		var options []oss.UploadOption

		if _key == "" {
			_key = _KODO_FILE_KEY_PREFIX_DEFAULT + "/" + _filename
		}
		options = append(options, kodo.WithKey(_key))
		if _expiry > 0 {
			options = append(options, kodo.WithExpiry(_expiry))
		}
		res, err := storage.Upload(ctx, fp, options...)
		if err != nil {
			utils.Fatal(err)
		}
		fmt.Printf("Saved file %s as %+v in kodo successfully!\n", _filename, res)
	default:
		panic("Unknown action: " + action)
	}

}
