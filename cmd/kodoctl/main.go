package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yuansl/playground/oss"
	"github.com/yuansl/playground/oss/kodo"
	"github.com/yuansl/playground/util"
)

const (
	_ACCESS_KEY_DEFAULT           = "557TpseUM8ovpfUhaw8gfa2DQ0104ZScM-BTIcBx"
	_SECRET_KEY_DEFAULT           = "d9xLPyreEG59pR01sRQcFywhm4huL-XEpHHcVa90"
	_KODO_BUCKET_DEFAULT          = "fusionlogtest"
	_KODO_FILE_KEY_PREFIX_DEFAULT = "share"
)

var _options struct {
	accessKey  string // = _ACCESS_KEY_DEFAULT
	secretKey  string // = _SECRET_KEY_DEFAULT
	bucket     string
	prefix     string
	limit      int
	filename   string
	key        string
	expiry     time.Duration
	output     string
	linkdomain string
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

func parseCmdArgs(args []string) {
	if len(args) == 0 {
		panic("BUG: args must not be empty")
	}
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.StringVar(&_options.linkdomain, "link", "http://pybwef48y.bkt.clouddn.com", "specify cdn domain of the download link")
	flags.StringVar(&_options.bucket, "bucket", _KODO_BUCKET_DEFAULT, "specify the kodo bucket")
	flags.StringVar(&_options.prefix, "prefix", _KODO_FILE_KEY_PREFIX_DEFAULT, "specify prefix of a object file")
	flags.IntVar(&_options.limit, "limit", 5, "list <limit> files at most")
	flags.StringVar(&_options.filename, "file", "", "specify a local file for uploading")
	flags.StringVar(&_options.key, "key", "", "specify file key in oss/kodo")
	flags.DurationVar(&_options.expiry, "expiry", 24*time.Hour, "specify expiry date of a file storead in oss")
	flags.StringVar(&_options.output, "o", "", "save as ...")
	if err := flags.Parse(args[1:]); err != nil {
		util.Fatal(err)
	}
}

func usage() {
	fmt.Printf(`kodoctl %s
Usage: %s [Command]

kodoctl is a command line tool for managing your files which stored in kodo

Command:
      download - download from kodo
      list     - list files stored in kodo
      upload   - upload a file to kodo
`, Version(), os.Args[0])

	os.Exit(0)
}

func main() {
	if len(os.Args) == 1 {
		usage()
	}
	parseCmdArgs(os.Args[1:])

	storage := kodo.NewStorageService(kodo.WithCredential(_options.accessKey, _options.secretKey), kodo.WithLinkDomain(_options.linkdomain))
	ctx := util.InitSignalHandler(context.TODO())

	switch action := os.Args[1]; action {
	case "list":
		var options []oss.ListOption
		if _options.prefix != "" {
			options = append(options, kodo.WithListPrefix(_options.prefix))
		}
		if _options.limit > 0 {
			options = append(options, kodo.WithListLimit(_options.limit))
		}
		files, err := storage.List(ctx, _options.bucket, options...)
		if err != nil {
			util.Fatal(err)
		}
		for _, it := range files {
			fmt.Printf("file: %+v\n", it)
		}
	case "download":
		if _options.key == "" {
			util.Fatal("kodo: oss key must not be empty")
		}
		data, err := storage.Download(ctx, _options.bucket, _options.key)
		if err != nil {
			util.Fatal("store.Download error:", err)
		}
		if _options.output == "" {
			_options.output = filepath.Base(_options.key)
		}
		output, err := os.OpenFile(_options.output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			util.Fatal("os.OpenFile:", err)
		}
		defer output.Close()

		if _, err = output.Write(data); err != nil {
			util.Fatal("os.Write:", err)
		}
	case "upload":
		path, _ := filepath.Abs(_options.filename)
		fp, err := os.Open(path)
		if err != nil {
			util.Fatal(err)
		}
		defer fp.Close()

		var options []oss.UploadOption

		if _options.key == "" {
			_options.key = _KODO_FILE_KEY_PREFIX_DEFAULT + "/" + filepath.Base(path)
		}
		options = append(options, kodo.WithKey(_options.key))
		if _options.expiry > 0 {
			options = append(options, kodo.WithExpiry(_options.expiry))
		}
		res, err := storage.Upload(ctx, _options.bucket, fp, options...)
		if err != nil {
			util.Fatal(err)
		}
		fmt.Printf("Saved file %s as %+v in kodo successfully!\n", _options.filename, res)
	case "-h", "--help":
		fallthrough
	default:
		usage()
	}
}
