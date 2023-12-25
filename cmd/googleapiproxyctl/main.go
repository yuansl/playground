package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/qbox/net-deftones/googleapisproxy"
	"github.com/qbox/net-deftones/googleapisproxy/proto"
	"github.com/qbox/net-deftones/logger"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"

	"github.com/yuansl/playground/util"
)

const scheme = "yuansl"

var _options struct {
	address   string
	enableTLS bool
}

func parseCmdArgs() {
	flag.StringVar(&_options.address, "address", "120.241.29.6:443", "specify addres(proxy)")
	flag.BoolVar(&_options.enableTLS, "tls", false, "if enable TLS support for grpc connection")
	flag.Parse()
}

// initGrpcResolver registers a customized Resolver for grpc connection
func initGrpcResolver() {
	resolver0 := manual.NewBuilderWithScheme(scheme)
	resolver0.InitialState(resolver.State{Addresses: []resolver.Address{{Addr: _options.address}}})
	resolver.Register(resolver0)
}

func main() {
	parseCmdArgs()
	initGrpcResolver()

	client, err := googleapisproxy.NewGcloudClient(scheme+"://defy-googleapisproxy.qiniu.com:443",
		googleapisproxy.WithCAFile("/etc/ssl/certs/ca-certificates.crt"),
		googleapisproxy.WithServerName("*.qiniuapi.com"))
	if err != nil {
		util.Fatal(err)
	}
	ctx := logger.NewContext(context.TODO(), logger.New())

	if result, err := client.ListUrlMaps(ctx, &proto.ListUrlMapsRequest{ProjectId: "qiniu-cdn"}); err != nil {
		util.Fatal(err)
	} else {
		fmt.Printf("urlmaps: %+v\n", result.Urlmaps)
	}
}
