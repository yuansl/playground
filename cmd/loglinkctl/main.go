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
	"strings"
	"time"

	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/util"
	"golang.org/x/sync/errgroup"

	"github.com/yuansl/playground/cmd/loglinkctl/repository"
	"github.com/yuansl/playground/cmd/loglinkctl/repository/fusionlog"
	"github.com/yuansl/playground/oss/kodo"
)

const _PRIVATE_URL_EXPIRY_DEFAULT = 2 * 31 * 24 * time.Hour // 2 month

var options struct {
	accessKey, secretKey string
	linkdomain           string
	mongouri             string
	domains              string
	bucket               string
	begin                time.Time
	end                  time.Time
	verbose              bool
	continueOnError      bool
}

func init() {
	options.accessKey = os.Getenv("ACCESS_KEY")
	options.secretKey = os.Getenv("SECRET_KEY")
}

func parseOptions() {
	flag.StringVar(&options.linkdomain, "linkdomain", "https://fusionlog.qiniu.com", "specify kodo link domain")
	flag.StringVar(&options.mongouri, "mongouri", "mongodb://127.0.0.1:27017/fusionlogv2", "specify mongo connect string")
	flag.TextVar(&options.begin, "begin", time.Time{}, "begin time (in RFC3339)")
	flag.TextVar(&options.end, "end", time.Time{}, "end time (in RFC3339)")
	flag.StringVar(&options.domains, "domains", "", "specify domain (seperate by comma)")
	flag.StringVar(&options.bucket, "bucket", "fusionlog", "specify (kodo)bucket name")
	flag.BoolVar(&options.verbose, "v", false, "verbose")
	flag.BoolVar(&options.continueOnError, "c", false, "whether continue or stop on error")
	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s -begin <begin> -end <end>\n", os.Args[0])
		os.Exit(1)
	}
	if options.begin.After(options.end) || options.end.IsZero() {
		util.Fatal("Invalid time range")
	}
}

func RefreshLoglinks(ctx context.Context, links []repository.LogLink, storage *kodo.StorageService, repo repository.LinkRepository) {
	for _, link := range links {
		newUrl := storage.UrlOfKey(ctx, options.bucket, link.Name, kodo.WithPrivateUrlExpiry(_PRIVATE_URL_EXPIRY_DEFAULT))
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

func Run(ctx context.Context, domains []string, begin, end time.Time, repo repository.LinkRepository) error {
	log := logger.FromContext(ctx)
	wg, ctx := errgroup.WithContext(ctx)

	for _, domain := range domains {
		wg.Go(func() error {
			links, err := repo.GetLinks(ctx, begin, end,
				&repository.LinkOptions{
					Domain:   domain,
					Business: repository.BusinessCdn,
					Filter: func(link *repository.LogLink) bool {
						return !strings.HasPrefix(link.Url, "http") && (begin.Compare(link.Timestamp) <= 0 && link.Timestamp.Before(end) && !strings.Contains(link.Name, "cztv"))
					},
				})
			if err != nil {
				return fmt.Errorf("repo.GetLinks: %v", err)
			}
			log.Infof("got %d log links between %s and %s of domain '%s'\n",
				len(links), begin, end, domain)

			for _, link := range links {
				wg.Go(func() error {
					log.Infof("%+v\n", link)

					/*
						err := InspectLogLink(ctx, &link)
						if err != nil {
							fmt.Print(err)
							if options._continue {
								return nil
							}
						}
					*/
					return nil
				})
			}
			return repo.DeleteLinks(ctx, links...)
		})
	}
	return wg.Wait()
}

func main() {
	parseOptions()

	repo, err := fusionlog.NewLinkRepository(fusionlog.NewFusionlogdb(options.mongouri))
	if err != nil {
		util.Fatal("NewLinkRepository:", err)
	}
	ctx := context.Background()
	if options.verbose {
		ctx = logger.NewContext(ctx, logger.New())
	}
	ctx = util.InitSignalHandler(ctx)

	err = Run(ctx, strings.Split(options.domains, ","), options.begin, options.end, repo)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			err = context.Cause(ctx)
		default:
		}
		util.Fatal("Run:", err)
	}
	fmt.Printf("DONE\n")
}
