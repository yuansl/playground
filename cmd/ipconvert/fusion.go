package main

import (
	"context"

	"github.com/qiniu/go-sdk/v7/cdn"
)

type fusion struct {
	*cdn.CdnManager
}

func (fusion *fusion) RefreshUrl(ctx context.Context, url string) error {
	_, err := fusion.CdnManager.RefreshUrls([]string{url})
	return err
}
