package main

import (
	"context"
	"net/http"

	"github.com/qbox/net-deftones/util/httputil"
)

func RegisterHttpHandlers(ctx context.Context, srv RaftService) http.Handler {
	handler := httputil.InitHttpHandlerRegister()

	handler.POST("/v1/getlog", func(c httputil.HttpContext) {
		var request GetlogRequest
		httputil.HandleRequest(c, &request, func(ctx context.Context) (any, error) {
			return srv.HandleGetlogRequest(ctx, &request)
		})
	})

	handler.POST("/v1/addlog", func(c httputil.HttpContext) {
		var request AddlogRequest
		httputil.HandleRequest(c, &request, func(ctx context.Context) (any, error) {
			return srv.HandleAddlogRequest(ctx, &request)
		})
	})
	handler.POST("/v1/appendlog", func(c httputil.HttpContext) {
		var request AppendLogRequest
		httputil.HandleRequest(c, &request, func(ctx context.Context) (any, error) {
			return srv.HandleAppendLogEntry(ctx, &request)
		})
	})
	handler.POST("/v1/vote", func(c httputil.HttpContext) {
		var request VoteRequest
		httputil.HandleRequest(c, &request, func(ctx context.Context) (any, error) {
			return srv.HandleVoteRequest(ctx, &request)
		})
	})

	return handler.(http.Handler)
}
