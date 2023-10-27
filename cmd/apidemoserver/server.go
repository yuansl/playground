package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/yuansl/playground/apidemo/api/rpc/proto"
	"github.com/yuansl/playground/logger"
	"github.com/yuansl/playground/tracer"
	"github.com/yuansl/playground/util"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/ratelimit"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const _UID_DEFAULT = 999999

type Server struct {
	proto.UnimplementedAccountServer
	grpcSrv     *grpcServer
	httpSrv     *http.Server
	config      *Config
	ratelimiter ratelimit.Limiter
}

func ProtoError(err error) error {
	s, _ := status.New(codes.InvalidArgument, codes.InvalidArgument.String()).WithDetails(&proto.Error{
		Code: 400001, Message: err.Error()})
	return s.Err()
}

func (srv *Server) ListUser(ctx context.Context, req *proto.UserRequest) (*proto.UserResponse, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	srv.ratelimiter.Take()

	if req.Uid != _UID_DEFAULT {
		return nil, ProtoError(fmt.Errorf("invalid uid: %d", req.Uid))
	}

	return &proto.UserResponse{Result: []*proto.User{
		{Name: "liming", Age: 21, Address: "Beijing, China", Id: span.SpanContext().SpanID().String()}},
	}, nil
}

func (srv *Server) Close() {
	srv.grpcSrv.Shutdown(context.TODO())
	srv.httpSrv.Shutdown(context.TODO())
}

func (srv *Server) Run(ctx context.Context) error {
	var errorq = make(chan error, 1)

	go func() {
		errorq <- srv.grpcSrv.ListenAndServe()
	}()

	go func() {
		errorq <- srv.httpSrv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		srv.Close()

		switch err := ctx.Err(); {
		case errors.Is(err, context.Canceled):
			return context.Cause(ctx)
		default:
			return err
		}
	case err := <-errorq:
		return err
	}
}

type grpcServer struct {
	*grpc.Server
	Addr string
}

func (srv *grpcServer) ListenAndServe() error {
	listener, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		util.Fatal(err)
	}
	return srv.Serve(listener)
}

func (srv *grpcServer) Shutdown(context.Context) error {
	srv.GracefulStop()

	return nil
}

func initializeGrpcServer(srv *Server) {
	srv.grpcSrv = &grpcServer{
		Server: grpc.NewServer(
			grpc.ConnectionTimeout(srv.config.GrpcServer.ConnectTimeout),
			grpc.MaxRecvMsgSize(srv.config.GrpcServer.MaxMsgSize),
			grpc.ChainUnaryInterceptor(
				otelgrpc.UnaryServerInterceptor(otelgrpc.WithTracerProvider(tracer.GetTracerProvider())),
				grpc.UnaryServerInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
					span := trace.SpanFromContext(ctx)
					ctx = logger.NewContext(ctx, logger.New(span.SpanContext().TraceID().String()))

					grpc.SendHeader(ctx, metadata.Pairs("X-Reqid", span.SpanContext().TraceID().String()))

					return handler(ctx, req)
				})),
		),
		Addr: srv.config.GrpcServer.Addr,
	}

	proto.RegisterAccountServer(srv.grpcSrv, srv)
}

func initializeHttpServer(srv *Server) {
	mux := runtime.NewServeMux(runtime.WithErrorHandler(runtime.ErrorHandlerFunc(func(ctx context.Context, _ *runtime.ServeMux, _ runtime.Marshaler, w http.ResponseWriter, req *http.Request, err error) {
		if md, ok := metadata.FromOutgoingContext(ctx); ok {
			if id, exists := md["x-reqid"]; exists {
				ctx = logger.NewContext(ctx, logger.New(id...))
			}
		}
		if s, ok := status.FromError(err); ok {
			if details := s.Proto().Details; len(details) > 0 {
				var cause proto.Error

				if err := details[0].UnmarshalTo(&cause); err != nil {
					panic("bug: anypb.Any.UnmarshalTo error: " + err.Error())
				}
				logger.FromContext(ctx).Infof("cause = '%+v'\n", &cause)
				data, _ := json.Marshal(&proto.UserResponse{Error: &cause})
				w.WriteHeader(http.StatusBadRequest)
				w.Write(data)
				return
			}
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": {"code": 500001, "message": "internal error", "data": ["unknown internal error, please contact your administrator"]}}`))
	})))

	srv.httpSrv = &http.Server{
		Addr:    srv.config.HttpServer.Addr,
		Handler: mux,
	}
	if err := http2.ConfigureServer(srv.httpSrv, nil); err != nil {
		panic("BUG: can't configure http/2 server")
	}

	cc, err := grpc.Dial(srv.config.GrpcServer.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}

	proto.RegisterAccountHandler(context.TODO(), mux, cc)
}

func NewServer(cfg *Config) *Server {
	srv := Server{config: cfg, ratelimiter: ratelimit.New(1)}

	initializeGrpcServer(&srv)
	initializeHttpServer(&srv)

	return &srv
}
