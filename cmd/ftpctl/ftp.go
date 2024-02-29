package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"time"

	"github.com/jlaffaye/ftp"
)

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
func (f *FtpFile) Size() int64 {
	return int64(f.ftp.Size)
}

// Sys implements fs.FileInfo.
func (f *FtpFile) Sys() any {
	return nil
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

func (cli *ftpclient) Rmdir(path string) error {
	return cli.conn.RemoveDir(path)
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

func (cli *ftpclient) Delete(path string) error {
	return cli.conn.Delete(path)
}

func (*ftpclient) Get(path string) ([]byte, error) {
	return nil, errors.ErrUnsupported
}

func (*ftpclient) Put(local io.Reader, remotepath string) error {
	return errors.ErrUnsupported
}

func (cli *ftpclient) Chdir(path string) error {
	return cli.conn.ChangeDir(path)
}

func (cli *ftpclient) Quit() error {
	return cli.conn.Quit()
}

func Login(addr, user, password string) *ftpclient {
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
