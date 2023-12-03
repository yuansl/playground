package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

type Person struct {
	Name    string
	Age     int
	Address string
}

type KV struct {
	perNameRecord map[string]Person
}

type KVRequest struct {
	Name string
}
type KVResponse struct {
	Result *Person `json:,omitempty`

	Error *struct {
		Code    int
		Message string
		Data    any
	} `json:",omitempty"`
}

func (kv *KV) Get(req *KVRequest, reply *KVResponse) error {
	if v, exist := kv.perNameRecord[req.Name]; exist {
		reply.Result = &v
	} else {
		return errors.New("KV: not found")
	}
	return nil
}

func (kv *KV) Set(req *KVRequest, reply *KVResponse) error {
	return nil
}

func main() {
	go startServer()

	time.Sleep(2 * time.Second)

	client, err := rpc.DialHTTP("tcp", ":8080")
	if err != nil {
		fatal("rpc.DialHttp error: %v\n", err)
	}

	var res KVResponse
	err = client.Call("KV.Get", &KVRequest{Name: "Liujie"}, &res)
	if err != nil {
		fatal("client.Call error: %v\n", err)
	}

	fmt.Printf("response: %+v\n", res.Result)
}

func startServer() {
	kv := KV{
		perNameRecord: map[string]Person{
			"Liujie": {
				Name:    "Liujie",
				Age:     26,
				Address: "Shanghai",
			},
		},
	}
	rpc.Register(&kv)
	rpc.HandleHTTP()

	fatal("last error: %v\n", http.ListenAndServe(":8080", nil))
}

func fatal(format string, v ...any) {
	fmt.Fprintf(os.Stderr, format, v...)
	os.Exit(1)
}
