package domaininfo

import (
	"errors"
	"strings"
)

type DomainInfo struct {
	Name     string
	Origin   string
	Uid      uint
	Conflict bool
}

const SUFFIX = ".fusion-conflict"

var ErrInvalidName = errors.New("Invalid domain name")

func transform(d *DomainInfo) error {
	idx := strings.LastIndex(d.Name, SUFFIX)
	if idx == -1 {
		return ErrInvalidName
	}
	idx = strings.LastIndex(d.Name[:idx], ".")
	if idx == -1 {
		return ErrInvalidName
	}
	d.Origin = d.Name
	d.Name = d.Name[:idx]
	return nil
}

func nameWithoutConflict(d *DomainInfo) error {
	if !d.Conflict {
		return nil
	}
	if !strings.HasSuffix(d.Name, SUFFIX) {
		return ErrInvalidName
	}
	parts := strings.Split(d.Name, ".")
	if len(parts) < 3 {
		return ErrInvalidName
	}
	d.Origin = d.Name
	d.Name = strings.Join(parts[:len(parts)-2], ".")
	return nil
}
