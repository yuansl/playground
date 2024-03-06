package main

import (
	"context"
	"net/http"

	"github.com/qbox/net-deftones/util/httputil"
)

func RegisterHttpHandlers(ctx context.Context, srv RaftService) http.Handler {
	handler := httputil.InitHttpHandlerRegister()

	handler.GET("/v1/appendlog", func(c httputil.HttpContext) {
		var request AppendLogRequest
		httputil.HandleRequest(c, &request, func(ctx context.Context) (any, error) {
			return srv.HandleAppendLogEntry(ctx, &request)
		})
	})
	handler.GET("/v1/vote", func(c httputil.HttpContext) {
		var request VoteRequest
		httputil.HandleRequest(c, &request, func(ctx context.Context) (any, error) {
			return srv.HandleVoteRequest(ctx, &request)
		})
	})

	return handler.(http.Handler)
}
