package httputil

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/util"
)

const _GIN_CONTEXT_KEY = "context"

func contextFromGin(c *gin.Context) context.Context {
	if _ctx, exists := c.Get(_GIN_CONTEXT_KEY); exists {
		return _ctx.(context.Context)
	}
	return c.Request.Context()
}

type Response struct {
	Result any `json:"result"`
	*Error `json:"error,omitempty"`
}

var EmptyResponse Response

type HttpContext = gin.Context

func HandleRequest(c *HttpContext, request any, handle func(ctx context.Context) (any, error)) {
	if err := c.ShouldBind(request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Response{
			Error: NewErrorWith(ErrInvalid, err),
		})
		return
	}
	ctx := contextFromGin(c)
	logger.FromContext(ctx).Infof("Request: %+v\n", request)

	res, err := handle(ctx)
	if err != nil {
		logger.FromContext(ctx).Warnf(
			"The request %q failed: %v\n", c.Request.RequestURI, err)

		var cause *Error

		if !errors.As(err, &cause) {
			panic("BUG: the type of err must be *httputil.Error")
		}
		if cause == nil {
			c.JSON(http.StatusOK, &EmptyResponse)
			return
		}

		c.AbortWithStatusJSON(cause.Code/1000, Response{Error: cause})
		return
	}

	c.JSON(http.StatusOK, &Response{Result: res})
}

func getRequestId(c *HttpContext) string {
	xreqid := c.Request.Header.Get("X-Reqid")

	if xreqid == "" {
		xreqid = logger.IdFromContext(contextFromGin(c))
	}
	return xreqid
}

type HandlerRegistry = gin.IRoutes

func InitHttpHandlerRegister() HandlerRegistry {
	gin.SetMode(gin.ReleaseMode)

	return gin.New().Use(func(c *gin.Context) {
		log := logger.NewWith(c.Request)
		ctx := logger.WithContext(c.Request.Context(), log)

		c.Set(_GIN_CONTEXT_KEY, ctx)

		if xreqId := getRequestId(c); xreqId != "" {
			c.Writer.Header().Set("X-Reqid", xreqId)
		}
		log.Infof("%s %s\n", c.Request.Method, c.Request.RequestURI)
	}, gin.Recovery())
}

func StartHttpServer(ctx context.Context, addr string, handler http.Handler) error {
	server := http.Server{
		Addr:    addr,
		Handler: handler,
	}
	defer server.Shutdown(ctx)

	return util.WithContext(ctx, func() error {
		return server.ListenAndServe()
	})
}
