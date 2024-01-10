package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/util"
	"golang.org/x/sync/errgroup"

	"github.com/yuansl/playground/cmd/loglinkctl/repository"
	"github.com/yuansl/playground/oss/kodo"
)

const _PRIVATE_URL_EXPIRY_DEFAULT = 2 * 31 * 24 * time.Hour // 2 month

var _options struct {
	accessKey, secretKey string
	linkdomain           string
	mongouri             string
	domain               string
	bucket               string
	begin                time.Time
	end                  time.Time
	verbose              bool
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
	flag.StringVar(&_options.bucket, "bucket", "fusionlog", "specify (kodo)bucket name")
	flag.BoolVar(&_options.verbose, "v", false, "verbose")
	flag.Parse()
}

func RefreshLoglinks(ctx context.Context, links []repository.LogLink, storage *kodo.StorageService, repo repository.LinkRepository) {
	for _, link := range links {
		newUrl := storage.UrlOfKey(ctx, _options.bucket, link.Name, kodo.WithPrivateUrlExpiry(_PRIVATE_URL_EXPIRY_DEFAULT))
		if err := repo.SetDownloadUrl(ctx, &link, newUrl); err != nil {
			util.Fatal(err)
		}
	}
}

func InspectLogLink(ctx context.Context, link *repository.LogLink) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, link.Url, nil)
	if err != nil {
		return fmt.Errorf("http.Head: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http.Client.Do: %w", err)
	}
	defer resp.Body.Close()

	data, _ := httputil.DumpResponse(resp, true)

	if lengthHeader := resp.Header.Get("Content-Length"); lengthHeader != "" {
		contentLen, err := strconv.Atoi(lengthHeader)
		if err != nil || contentLen != int(link.Size) {
			return fmt.Errorf("mismatch: expected size: %d, but got: %s, link: %+v", link.Size, lengthHeader, link)
		}
	} else {
		return fmt.Errorf("mismatch: bad response(raw): '%s', 'Content-Length' header missing, link: %+v", data, link)
	}
	return nil
}

func Run(ctx context.Context, repo repository.LinkRepository) error {
	return util.WithContext(ctx, func() error {
		links, err := repo.GetLinks(ctx, _options.domain, _options.begin, _options.end, repository.BusinessCdn)
		if err != nil {
			return fmt.Errorf("repo.GetLinks: %v", err)
		}
		log := logger.FromContext(ctx)
		log.Infof("got %d log links between %s and %s of domain '%s'\n",
			len(links), _options.begin, _options.end, _options.domain)

		var counter atomic.Int32
		wg, ctx := errgroup.WithContext(ctx)
		go func() {
			for range time.Tick(1 * time.Second) {
				log.Infof("inspected %d link\n", counter.Load())
			}
		}()
		for _, link := range links {
			_link := link
			wg.Go(func() error {
				defer counter.Add(+1)
				log.Infof("Inspecting link %+v ...\n", _link)

				return InspectLogLink(ctx, &_link)
			})
		}
		return wg.Wait()
	})
}

func main() {
	parseCmdOptions()

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s -begin <begin> -end <end>\n", os.Args[0])
		os.Exit(1)
	}
	if _options.begin.After(_options.end) || _options.end.IsZero() {
		util.Fatal("Invalid time range")
	}
	if _options.secretKey == "" || _options.accessKey == "" {
		util.Fatal("either access key or secret key must not be empty")
	}

	// storage := kodo.NewStorageService(kodo.WithCredential(_options.accessKey, _options.secretKey), kodo.WithLinkDomain(_options.linkdomain))
	repo, err := repository.NewMongoLinkRepository(_options.mongouri)
	if err != nil {
		util.Fatal("NewMongoLinkRepository:", err)
	}
	ctx := context.Background()

	if _options.verbose {
		ctx = logger.NewContext(ctx, logger.New())
	}
	ctx = util.InitSignalHandler(ctx)

	err = Run(ctx, repo)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			err = context.Cause(ctx)
		}
		util.Fatal("Run:", err)
	}
	fmt.Printf("DONE\n")
}
