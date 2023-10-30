package kodo

import (
	"crypto/md5"
	"encoding/hex"
	"hash"
	"io"
)

type md5InterceptReader struct {
	r         io.Reader
	md5digest hash.Hash
}

func (md5rw *md5InterceptReader) Read(p []byte) (int, error) {
	n, err := md5rw.r.Read(p)
	if err != nil {
		return n, err
	}
	md5rw.md5digest.Write(p[:n])

	return n, nil
}

func (md5rw *md5InterceptReader) Sum() string {
	return hex.EncodeToString(md5rw.md5digest.Sum(nil))
}

func NewMd5InterceptReader(r io.Reader) *md5InterceptReader {
	return &md5InterceptReader{md5digest: md5.New(), r: r}
}
