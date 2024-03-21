// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-12-25 13:29:18

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
	"fmt"
	"io"
	"io/fs"
	"log"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/qbox/net-deftones/stream"
	"github.com/qbox/net-deftones/util"
)

var options struct {
	connectTimeout time.Duration
	ftpaddr        string
	user           string
	password       string
	startkey       string
	endkey         string
}

func parseCmdOptions() {
	flag.StringVar(&options.ftpaddr, "addr", "115.238.46.167:21", "specify address of ftp server")
	flag.StringVar(&options.user, "user", "qiniup", "ftp user")
	flag.StringVar(&options.password, "passwd", "9Z3AMF3Q", "ftp password")
	flag.DurationVar(&options.connectTimeout, "timeout", 1800*time.Second, "specify connect timeout in seconds")
	flag.StringVar(&options.startkey, "start", "a", "start key")
	flag.StringVar(&options.endkey, "end", "z", "end key")
	flag.Parse()
}

type FTPService interface {
	Chdir(path string) error
	Put(local io.Reader, remotepath string) error
	Get(path string) ([]byte, error)
	List(path ...string) ([]fs.DirEntry, error)
	Delete(path string) error
	Mkdir(path string) error
	Rmdir(path string) error
	Quit() error
}

func WalkFTPDirectory(dirent fs.DirEntry, deleteWalk bool, ftp FTPService, callbackf func(fs.DirEntry)) {
	// fmt.Printf("cd %s\n", dirent.Name())
	err := ftp.Chdir(dirent.Name())
	if err != nil {
		util.Fatal("cd %s failed: %v\n", dirent.Name(), err)
	}
	ents, err := ftp.List()
	if err != nil {
		util.Fatal("ls failed: %v\n", err)
	}
	for _, ent := range ents {
		if ent.IsDir() {
			if strings.HasPrefix(ent.Name(), "2024-02") {
				WalkFTPDirectory(ent, deleteWalk, ftp, callbackf)
			}
		} else {
			callbackf(ent)
		}
	}
	if err := ftp.Chdir("../"); err != nil {
		util.Fatal("cd ../ failed: %v\n", err)
	}

	if deleteWalk {
		fmt.Printf("rmdir %s\n", dirent.Name())
		if err := ftp.Rmdir(dirent.Name()); err != nil {
			fmt.Printf("rmdir %s failed: %v\n", dirent.Name(), err)
		}
	}
}

type DomainTime struct {
	Domain string
	time.Time
}

func deleteWalkDir(dirent fs.DirEntry) bool {
	return !strings.HasSuffix(dirent.Name(), ".cztv.com") && !strings.Contains(dirent.Name(), "cztv")
}

const _POINTS_PER_DAY = 24 * 12
const _POINTS_PER_MONTH = 29 * _POINTS_PER_DAY

func aggregate(domaintimes *stream.Pair[string, any], missedhours *util.Set[DomainTime]) {
	times := domaintimes.Value.([]DomainTime)
	if llen := len(times); llen < _POINTS_PER_MONTH {
		fmt.Printf("domain:%s missing some files: expected %d, actual: %d\n", domaintimes.Key, _POINTS_PER_MONTH, llen)

		filecountof := func(alignas time.Duration) []stream.Pair[time.Time, any] {
			count := stream.StreamOf[DomainTime, time.Time, int](times).
				GroupBy(func(t DomainTime) time.Time {
					return alignoftime(t.Time, alignas).Local()
				}).
				ReduceByKey(
					func(pair stream.Pair[time.Time, any]) stream.Pair[time.Time, []int] {
						var counts = make([]int, 0, len(pair.Value.([]time.Ticker)))
						for range pair.Value.([]time.Time) {
							counts = append(counts, +1)
						}
						return stream.Pair[time.Time, []int]{Key: pair.Key, Value: counts}
					},
					stream.BinaryOpFn[int](func(a, b int) int { return a + b }),
				).
				Collect() // []{{time.Time -> per-time-counter}, ...}

			sort.Slice(count, func(i, j int) bool {
				return count[i].Key.Before(count[j].Key)
			})
			return count
		}

		perdayfilecount := filecountof(24 * time.Hour)

		start := time.Date(2024, 2, 1, 0, 0, 0, 0, time.Local)
		end := start.AddDate(0, +1, 0)
		sameday := func(t1, t2 time.Time) bool {
			return t1.Year() == t2.Year() && t2.Month() == t2.Month() && t1.Day() == t2.Day()
		}
		for day := start; day.Before(end); day = day.Add(24 * time.Hour) {
			index := slices.IndexFunc(perdayfilecount, func(e stream.Pair[time.Time, any]) bool { return sameday(e.Key, day) })
			if index == -1 || perdayfilecount[index].Value.(int) != _POINTS_PER_DAY {
				if index == -1 {
					for hour := day; hour.Before(day.Add(24 * time.Hour)); hour = hour.Add(time.Hour) {
						missedhours.Add(DomainTime{domaintimes.Key, hour})
					}
				} else {
					perhourfilecount := filecountof(time.Hour)

					for hour := day; hour.Before(day.Add(24 * time.Hour)); hour = hour.Add(time.Hour) {
						index := slices.IndexFunc(perdayfilecount, func(e stream.Pair[time.Time, any]) bool {
							return sameday(e.Key, day) && e.Key.Hour() == hour.Hour()
						})
						if index == -1 || perhourfilecount[index].Value.(int) != 12 {
							missedhours.Add(DomainTime{Domain: domaintimes.Key, Time: hour})
						}
					}
				}
			}
		}
	}
}

func stat(domains *util.Set[DomainTime]) {
	missedhours := util.NewSet[DomainTime]()

	stream.StreamOf[DomainTime, string, time.Time](domains.List()).
		GroupBy(func(d DomainTime) string {
			return d.Domain
		}). // []{{Domain -> []DomainTime}}
	ForEach(stream.ConsumerFn[*stream.Pair[string, any]](
		func(pair *stream.Pair[string, any]) { // Domain -> []DomainTime
			times := pair.Value.([]DomainTime)
			sort.Slice(times, func(i, j int) bool { return times[i].Time.Before(times[j].Time) })
			pair.Value = times

			aggregate(pair, missedhours)
		}))

	for _, it := range missedhours.List() {
		fmt.Printf("%s,%v\n", it.Domain, it.Time.Format("2006-01-02T15:04"))
	}
}

func run(_ context.Context, ftp FTPService) {
	domains := util.NewSet[DomainTime]()
	ents, err := ftp.List()
	if err != nil {
		log.Fatal("ftpcli.List:", err)
	}
	filter := func(dirent fs.DirEntry) bool {
		return strings.Contains(dirent.Name(), "cztv")
	}
	for _, dirent := range ents {
		if dirent.IsDir() && filter(dirent) {
			WalkFTPDirectory(dirent, deleteWalkDir(dirent), ftp, func(file fs.DirEntry) {
				// fmt.Printf("file %s\n", file.Name())

				if strings.HasSuffix(file.Name(), ".md5sum") {
					return
				}
				timestamp, err := time.ParseInLocation("2006-01-02_15:04", file.Name()[:16], time.Local)
				if err != nil {
					util.Fatal("invalid time parnttern: %q", file.Name())
				}
				fixedTime := alignoftime(timestamp, 5*time.Minute)

				domains.Add(DomainTime{Domain: dirent.Name(), Time: fixedTime.Local()})

				if deleteWalkDir(dirent) {
					fmt.Printf("delete %s\n", file.Name())
					if err := ftp.Delete(file.Name()); err != nil {
						util.Fatal("delete %s failed: %v\n", dirent.Name(), err)
					}
				}
			})
		}
	}
	stat(domains)
}

func main() {
	parseCmdOptions()

	ftp := Login(options.ftpaddr, options.user, options.password, options.connectTimeout)
	defer ftp.Quit()

	run(context.TODO(), ftp)
}
