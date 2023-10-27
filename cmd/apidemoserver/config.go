package main

import "time"

type GrpcServerConfig struct {
	Addr           string
	ConnectTimeout time.Duration
	MaxMsgSize     int
}

type HttpServerConfig struct {
	Addr string
}

type Config struct {
	GrpcServer GrpcServerConfig
	HttpServer HttpServerConfig
}

func initializeConfig() *Config {
	return &Config{
		GrpcServer: GrpcServerConfig{
			Addr:           ":10010",
			ConnectTimeout: 10 * time.Second,
			MaxMsgSize:     10_000_000,
		},
		HttpServer: HttpServerConfig{Addr: ":8080"},
	}
}
