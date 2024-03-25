// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-02-27 23:07:50

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
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/util"
	"github.com/qbox/net-deftones/util/httputil"
)

//go:generate stringer -type State -linecomment
type State int

const (
	FOLLOWER_STATE  State = iota // FOLLOWER
	CANDIDATE_STATE              // CANDIDATE
	LEADER_STATE                 // LEADER
)

var ErrInvalidState = errors.New("myraft: Invalid state")

type GetlogRequest struct {
	Key string
}

type GetlogResponse struct {
	Value string
}

type AddlogRequest struct {
	Key   string
	Value string
}

type AddlogResponse struct {
	Accept bool `json:"accept"`
}

type VoteRequest struct {
	Term int
	From string
}

type VoteResponse struct {
	Accept bool `json:"accept"`
}

type LogEntry struct {
	Term  int
	Key   string
	Value string
}

type AppendLogRequest struct {
	Term     int
	Commited bool
	Value    *LogEntry
	Name     string
}

type AppendLogResponse struct {
	Accept bool `json:"accept"`
}

type RaftService interface {
	HandleAppendLogEntry(ctx context.Context, req *AppendLogRequest) (*AppendLogResponse, error)
	HandleVoteRequest(ctx context.Context, req *VoteRequest) (*VoteResponse, error)
	HandleAddlogRequest(ctx context.Context, req *AddlogRequest) (*AddlogResponse, error)
	HandleGetlogRequest(ctx context.Context, req *GetlogRequest) (*GetlogResponse, error)
}

var options struct {
	addr    string
	members string
}

func parseOptions() {
	flag.StringVar(&options.addr, "addr", "127.0.0.1:3678", "specify the address of the node")
	flag.StringVar(&options.members, "members", "", "specify the members of the cluster")
	flag.Parse()

	fmt.Printf("options=%+v\n", options)

	if options.addr == "" {
		util.Fatal("addr of the server must not be empty")
	}
	if options.members == "" {
		util.Fatal("members must not be empty")
	}
}

func initClusterMembers(addrs []string) []*Node {
	var nodes []*Node

	for _, memb := range addrs {
		if strings.Contains(memb, options.addr) {
			continue
		}
		nodes = append(nodes, &Node{
			Addr: memb,
			Client: &httpClient{
				endpoint: memb, Client: http.DefaultClient,
			}})
	}
	return nodes
}

func Run(ctx context.Context, srv *RaftServer) error {
	handler := RegisterHttpHandlers(ctx, srv)

	go util.WithContext(ctx, func() error {
		return httputil.StartHttpServer(ctx, srv.Addr, handler)
	})
	return srv.TryRaftLeaderElection(ctx)
}

func main() {
	parseOptions()
	nodes := initClusterMembers(strings.Split(options.members, ","))
	server := RaftServer{Addr: options.addr, Others: nodes, Term: 0, election: time.NewTimer(electionTimeout())}
	ctx := util.InitSignalHandler(context.TODO())

	ctx = logger.NewContext(ctx, logger.New())

	if err := Run(ctx, &server); err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			err = context.Cause(ctx)
			fallthrough
		default:
			util.Fatal(err)
		}
	}
}
