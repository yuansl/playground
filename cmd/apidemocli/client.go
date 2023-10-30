package main

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/yuansl/playground/apidemo/api/rpc/proto"
	"github.com/yuansl/playground/trace"
)

var (
	ErrCanceled     = errors.New("account: Client canceled")
	ErrInvalid      = errors.New("account: Invalid argument")
	ErrTimeout      = errors.New("account: Request timeout")
	ErrRatelimit    = errors.New("account: Too many request")
	ErrUnauthorized = errors.New("account: Unauthorized")
	ErrUnknown      = errors.New("account: Unknown error")
)

func ErrorFrom(status *status.Status) error {
	var err error
	switch status.Code() {
	case codes.OK:
		return nil
	case codes.Canceled:
		err = ErrCanceled
	case codes.DeadlineExceeded:
		err = ErrTimeout
	case codes.InvalidArgument:
		err = ErrInvalid
	case codes.Unauthenticated:
		err = ErrUnauthorized
	case codes.ResourceExhausted:
		err = ErrRatelimit
	case codes.Unknown:
		fallthrough
	default:
		err = ErrUnknown
	}
	return fmt.Errorf("%w: %q", err, status.String())
}

type AccountService interface {
	ListUser(ctx context.Context, uid ...uint) ([]User, error)
}

type User struct {
	Name string
	Age  int
	ID   string
}

func proto_User2User(us []*proto.User) []User {
	var users []User
	for _, u := range us {
		users = append(users, User{Name: u.Name, Age: int(u.Age), ID: u.Id})
	}
	return users
}

type accountService struct {
	client        proto.AccountClient
	closeGrpcConn func() error
}

func (srv *accountService) ListUser(ctx context.Context, uid ...uint) ([]User, error) {
	ctx, span := trace.GetTracerProvider().Tracer("").Start(ctx, "accountService.ListUser")
	defer span.End()

	span.AddEvent(fmt.Sprintf("Request uid= '%v'", uid))

	var req proto.UserRequest
	if len(uid) > 0 {
		req.Uid = uint64(uid[0])
	}
	res, err := srv.client.ListUser(ctx, &req)
	if err != nil {
		if status, ok := status.FromError(err); ok {
			return nil, ErrorFrom(status)
		}
		return nil, ErrUnknown
	}
	return proto_User2User(res.Result), nil
}

func (cli *accountService) Close() {
	cli.closeGrpcConn()
	trace.GetTracerProvider().Shutdown(context.TODO())
}

func NewAccountService(addr string) AccountService {
	grpcConn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor(otelgrpc.WithMessageEvents(otelgrpc.SentEvents))),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
	if err != nil {
		panic("BUG: grpc.Dial error:" + err.Error())
	}

	return &accountService{
		client:        proto.NewAccountClient(grpcConn),
		closeGrpcConn: grpcConn.Close,
	}
}
