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
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"time"

	"github.com/jlaffaye/ftp"
)

var options struct {
	connectTimeout time.Duration
	ftpaddr        string
	user           string
	password       string
}

func parseCmdOptions() {
	flag.StringVar(&options.ftpaddr, "addr", "115.238.46.167:21", "specify address of ftp server")
	flag.StringVar(&options.user, "user", "qiniup", "ftp user")
	flag.StringVar(&options.password, "passwd", "9Z3AMF3Q", "ftp password")
	flag.DurationVar(&options.connectTimeout, "timeout", 50*time.Second, "specify connect timeout in seconds")
	flag.Parse()
}

func listFTPServer(ftpcli *ftp.ServerConn, path string) []*ftp.Entry {
	ents, err := ftpcli.List(path)
	if err != nil {
		log.Fatal("😡😤 ftp.List: ", err)
	}
	for _, ent := range ents {
		ent.Time = ent.Time.Local()
		fmt.Printf("  file entry: '%+v' 🍺🎉\n", ent)
	}
	return ents
}

type ftpclient struct {
	conn *ftp.ServerConn
}

type FtpFile struct{ *DirEntry }

// ModTime implements fs.FileInfo.
func (f *FtpFile) ModTime() time.Time {
	return f.ftp.Time
}

// Mode implements fs.FileInfo.
func (f *FtpFile) Mode() fs.FileMode {
	return f.DirEntry.Type()
}

// Size implements fs.FileInfo.
func (*FtpFile) Size() int64 {
	panic("unimplemented")
}

// Sys implements fs.FileInfo.
func (*FtpFile) Sys() any {
	panic("unimplemented")
}

var _ fs.FileInfo = (*FtpFile)(nil)

type DirEntry struct {
	ftp  *ftp.Entry
	mode fs.FileMode
}

// Info implements fs.DirEntry.
func (ent *DirEntry) Info() (fs.FileInfo, error) {
	return &FtpFile{DirEntry: ent}, nil
}

// IsDir implements fs.DirEntry.
func (ent *DirEntry) IsDir() bool {
	return ent.Type().IsDir()
}

// Name implements fs.DirEntry.
func (ent *DirEntry) Name() string {
	return ent.ftp.Name
}

// Type implements fs.DirEntry.
func (ent *DirEntry) Type() fs.FileMode {
	return ent.mode
}

var _ fs.DirEntry = (*DirEntry)(nil)

func NewDirEntry(ent *ftp.Entry) *DirEntry {
	var mode fs.FileMode
	switch ent.Type {
	case ftp.EntryTypeFile:
		mode |= fs.ModeIrregular
	case ftp.EntryTypeFolder:
		mode |= fs.ModeDir
	case ftp.EntryTypeLink:
		mode |= fs.ModeSymlink
	default:
	}
	return &DirEntry{ftp: ent, mode: mode}
}

func (cli *ftpclient) List(path ...string) ([]fs.DirEntry, error) {
	var dirents []fs.DirEntry
	var dir string

	if len(path) == 0 {
		dir = "./"
	} else {
		dir = path[0]
	}
	ents, err := cli.conn.List(dir)
	if err != nil {
		return nil, fmt.Errorf("List: %v", err)
	}
	for _, ent := range ents {
		dirents = append(dirents, NewDirEntry(ent))
	}
	return dirents, nil
}

func (*ftpclient) Mkdir(path string) error {
	return errors.ErrUnsupported
}

func (*ftpclient) Delete(path string) error {
	return errors.ErrUnsupported
}

func (*ftpclient) Get(path string) ([]byte, error) {
	return nil, errors.ErrUnsupported
}

func (*ftpclient) Put(local io.Reader, remotepath string) error {
	return errors.ErrUnsupported
}

func (cli *ftpclient) Cd(path string) error {
	return cli.conn.ChangeDir(path)
}

func (cli *ftpclient) Close() error {
	return cli.conn.Quit()
}

func Connect(addr, user, password string) *ftpclient {
	serverConn, err := ftp.Dial(options.ftpaddr,
		ftp.DialWithTimeout(options.connectTimeout),
		ftp.DialWithDisabledEPSV(true),
		ftp.DialWithDialFunc(func(network, address string) (net.Conn, error) {
			log.Printf("connecting to %s (timeout=%s) ...\n", address, options.connectTimeout)
			return net.DialTimeout(network, address, options.connectTimeout)
		}))
	if err != nil {
		log.Fatal("ftp.Dial:", err)
	}

	if err = serverConn.Login(options.user, options.password); err != nil {
		log.Fatal("ftp.Login:", err)
	}

	return &ftpclient{conn: serverConn}
}

func prettyPrintFtpEntry(ents []fs.DirEntry) {
	for _, ent := range ents {
		fmt.Printf("entry: '%+v'\n", fs.FormatDirEntry(ent))
	}
}

func main() {
	parseCmdOptions()

	ftpcli := Connect(options.ftpaddr, options.user, options.password)
	defer ftpcli.Close()

	ents, err := ftpcli.List("/")
	if err != nil {
		log.Fatal("ftpcli.List:", err)
	}
	prettyPrintFtpEntry(ents)
	if len(ents) > 0 {
		if err = ftpcli.Cd("/qiniup-v.cztv.com/2023-12-25"); err != nil {
			log.Fatalf("ftp: cd: %v", err)
		}
		ents, err = ftpcli.List("./")
		if err != nil {
			log.Fatal("ftpcli.List:", err)
		}
		prettyPrintFtpEntry(ents)
	}
}
