package rpc

//go:generate protoc -I ./proto --go_opt=paths=source_relative --go_out=./proto --go-grpc_opt=paths=source_relative --go-grpc_out=./proto --grpc-gateway_out=./proto --grpc-gateway_opt=paths=source_relative --experimental_allow_proto3_optional ./proto/service.proto
