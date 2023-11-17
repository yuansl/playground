// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-11-15 16:14:51

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"context"
	"flag"
	"runtime"
	"time"

	"github.com/yuansl/playground/logger"
	"golang.org/x/sync/errgroup"
)

var (
	_begin       time.Time
	_end         time.Time
	_mongouri    string
	_db          string
	_collection  string
	_concurrency int
)

func parseCmdArgs() {
	flag.TextVar(&_begin, "begin", time.Time{}, "begin time (in RFC3339)")
	flag.TextVar(&_end, "end", time.Time{}, "end time (in RFC3339)")
	flag.StringVar(&_mongouri, "mongouri", "mongodb://xs1597:17002/log-validation", "specify mongouri")
	flag.StringVar(&_db, "db", "log-validation", "specify database name of log-validation service")
	flag.StringVar(&_collection, "coll", "domain_retry", "specify collection name of log-validation database")
	flag.IntVar(&_concurrency, "c", runtime.NumCPU(), "concurrency")
	flag.Parse()
}

type LogsysRepository interface {
	LogValidationRepository
	LogEtlRepository
}

func NewLogsysRepository(validationRepo LogValidationRepository, etlRepo LogEtlRepository) LogsysRepository {
	return struct {
		LogValidationRepository
		LogEtlRepository
	}{
		LogValidationRepository: validationRepo,
		LogEtlRepository:        etlRepo,
	}
}

func run(ctx context.Context, repo LogsysRepository) {
	egroup, ctx := errgroup.WithContext(ctx)

	for day := _begin; day.Before(_end); day = day.AddDate(0, 0, +1) {
		_day, _day_next := day, day.AddDate(0, 0, +1)
		// infos, err := repo.FindEtlTasks(ctx, _day, _day_next)
		// if err != nil {
		// 	util.Fatal(err)
		// }

		// fmt.Printf("got %d log-validation infos of day %v\n", len(infos), day)

		egroup.Go(func() error {
			return repo.DeleteEtlTasks(ctx, _day, _day_next)
		})
	}
	egroup.Wait()

}

func main() {
	parseCmdArgs()

	logValidationRepo := NewLogValidationRepository(NewDomainRetry(_mongouri, _db, _collection))
	logEtlRepo := NewLogEtlRepository(NewTaskdb(_mongouri, "logsys-etl", "raw_task_prod"))
	repo := NewLogsysRepository(logValidationRepo, logEtlRepo)
	ctx := logger.NewContext(context.TODO(), logger.New())

	run(ctx, repo)
}
