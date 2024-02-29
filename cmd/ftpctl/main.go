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
	"strings"
	"time"

	"github.com/yuansl/playground/util"
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

type FTPClient interface {
	Chdir(path string) error
	Put(local io.Reader, remotepath string) error
	Get(path string) ([]byte, error)
	List(path ...string) ([]fs.DirEntry, error)
	Delete(path string) error
	Mkdir(path string) error
	Rmdir(path string) error
	Quit() error
}

func WalkFTPDirectory(dirent fs.DirEntry, ftp FTPClient, callbackf func(fs.DirEntry)) {
	fmt.Printf("cd %s\n", dirent.Name())
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
			WalkFTPDirectory(ent, ftp, callbackf)
		} else {
			callbackf(ent)
		}
	}
	if err := ftp.Chdir("../"); err != nil {
		util.Fatal("cd ../ failed: %v\n", err)
	}
	fmt.Printf("rmdir %s\n", dirent.Name())
	if err := ftp.Rmdir(dirent.Name()); err != nil {
		fmt.Printf("rmdir %s failed: %v\n", dirent.Name(), err)
	}
}

type DomainTime struct {
	Domain string
	time.Time
}

func run(_ context.Context, ftp FTPClient) {
	// var perDomainFile = make(map[DomainTime]struct{})
	ents, err := ftp.List()
	if err != nil {
		log.Fatal("ftpcli.List:", err)
	}
	for _, dirent := range ents {
		if dirent.IsDir() && !strings.HasSuffix(dirent.Name(), ".cztv.com") {
			WalkFTPDirectory(dirent, ftp, func(file fs.DirEntry) {
				// fmt.Printf("file %s\n", entfile.Name())
				// timestamp, err := time.Parse("2006-01-02_15:04", file.Name()[:16])
				// if err != nil {
				// 	util.Fatal("invalid time parnttern: %q", file.Name())
				// }

				// fixedTime := time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), timestamp.Hour(), 0, 0, 0, timestamp.Location())
				// perDomainFile[DomainTime{Domain: dirent.Name(), Time: fixedTime}] = struct{}{}

				fmt.Printf("delete %s\n", file.Name())
				if err := ftp.Delete(file.Name()); err != nil {
					util.Fatal("delete %s failed: %v\n", dirent.Name(), err)
				}
			})
		}

	}
	// var perDomainTimes = make(map[string][]time.Time)
	// for key := range perDomainFile {
	// 	perDomainTimes[key.Domain] = append(perDomainTimes[key.Domain], key.Time)
	// }
	// for domain, times := range perDomainTimes {
	// 	sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })
	// 	perDomainTimes[domain] = times
	// }

	// for domain, times := range perDomainTimes {
	// 	fmt.Printf("domain:%s, timestamps:%v\n", domain, times)
	// }
}

func main() {
	parseCmdOptions()

	ftp := Login(options.ftpaddr, options.user, options.password)
	defer ftp.Quit()

	run(context.TODO(), ftp)
}
