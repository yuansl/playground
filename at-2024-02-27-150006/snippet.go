// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-02-27 15:00:06

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/smtp"
	"time"

	"github.com/yuansl/playground/util"
)

var ctxAuthKey int = 1

func AuthWithContext(ctx context.Context, auth smtp.Auth) context.Context {
	return context.WithValue(ctx, ctxAuthKey, auth)
}

func AuthFromContext(ctx context.Context) smtp.Auth {
	v := ctx.Value(ctxAuthKey)
	if auth, ok := v.(smtp.Auth); ok {
		return auth
	}
	return nil
}

func SendEmail(ctx context.Context, from string, to []string, msg []byte) error {
	var errorq = make(chan error, 1)
	go func() {
		auth := AuthFromContext(ctx)
		err := smtp.SendMail("smtphz.qiye.163.com:465", auth, from, to, msg)
		if err != nil {
			switch {
			case errors.Is(err, io.EOF):
				errorq <- SendEmail(ctx, from, to, msg)
				return
			default:
				errorq <- fmt.Errorf("smtp.SendMail: %w", err)
				return
			}
		}
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errorq:
		return err
	}
}

func main() {
	auth := smtp.PlainAuth("", "cdn-support@qiniu.com", "uAt7AQFBe6y8", "smtp.qiye.163.com")
	ctx := AuthWithContext(context.Background(), auth)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err := SendEmail(ctx, "cdn-support@qiniu.com", []string{"yuanshenglong@qiniu.com"}, []byte(`Hello, this is a test message`))
	if err != nil {
		util.Fatal("SendEmail:", err)
	}
}
