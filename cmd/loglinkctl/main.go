package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/util"

	"github.com/yuansl/playground/cmd/loglinkctl/repository"
	"github.com/yuansl/playground/oss/kodo"
)

const _PRIVATE_URL_EXPIRY_DEFAULT = 2 * 31 * 24 * time.Hour // 2 month

var _options struct {
	accessKey, secretKey string
	linkdomain           string
	mongouri             string
	domain               string
	begin                time.Time
	end                  time.Time
}

func init() {
	_options.accessKey = os.Getenv("ACCESS_KEY")
	_options.secretKey = os.Getenv("SECRET_KEY")
}

func parseCmdOptions() {
	flag.StringVar(&_options.linkdomain, "linkdomain", "https://fusionlog.qiniu.com", "specify kodo link domain")
	flag.StringVar(&_options.mongouri, "mongouri", "mongodb://127.0.0.1:27017", "specify mongo connect string")
	flag.TextVar(&_options.begin, "begin", time.Time{}, "begin time (in RFC3339)")
	flag.TextVar(&_options.end, "end", time.Time{}, "end time (in RFC3339)")
	flag.StringVar(&_options.domain, "domain", "", "specify domain")
	flag.Parse()
}

func main() {
	parseCmdOptions()

	if _options.begin.After(_options.end) || _options.end.IsZero() {
		fmt.Fprintf(os.Stderr, "Usage: %s -begin <begin> -end <end>", os.Args[0])
		os.Exit(1)
	}
	if _options.domain == "" {
		fmt.Fprintf(os.Stderr, "domain must not be empty")
		os.Exit(2)
	}
	if _options.secretKey == "" || _options.accessKey == "" {
		fmt.Fprintf(os.Stderr, "either access key or secret key must not be empty")
		os.Exit(3)
	}

	storage := kodo.NewStorageService(kodo.WithCredential(_options.accessKey, _options.secretKey), kodo.WithLinkDomain(_options.linkdomain))
	repo, err := repository.NewMongoLinkRepository(_options.mongouri)
	if err != nil {
		util.Fatal(err)
	}
	ctx := util.InitSignalHandler(logger.NewContext(context.TODO(), logger.New()))

	log := logger.FromContext(ctx)

	links, err := repo.GetLinks(ctx, _options.domain, _options.begin, _options.end, repository.BusinessSrc)
	if err != nil {
		util.Fatal(err)
	}
	log.Infof("got %d log links\n", len(links))
	for _, link := range links {

		newUrl := storage.UrlOfKey(ctx, "fusionlog", link.Name, kodo.WithPrivateUrlExpiry(_PRIVATE_URL_EXPIRY_DEFAULT))
		if err = repo.SetDownloadUrl(ctx, &link, newUrl); err != nil {
			util.Fatal(err)
		}
	}
}
