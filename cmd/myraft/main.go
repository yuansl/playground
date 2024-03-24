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
	"sync"
	"time"

	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/util"
	"github.com/qbox/net-deftones/util/httputil"
)

type State int

const (
	FOLLOWER_STATE State = iota
	CANDIDATE_STATE
	LEADER_STATE
)

var ErrInvalidState = errors.New("myraft: Invalid state")

type VoteRequest struct {
	Term int
	From string
}

type VoteResponse struct {
	Accept bool `json:"accept"`
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
	// SendAppendLogRequest(req *AppendLogRequest) (*AppendLogResponse, error)
}

type LogEntry struct {
	Term  int
	Value any
}

type Server struct {
	Addr          string
	Leader        string
	mu            sync.RWMutex
	state         State
	Term          int
	Others        []*Node
	voteFor       string
	election      *time.Timer
	healthcheck   *time.Ticker
	appendlogChan chan *AppendLogRequest
	commited      []LogEntry
	uncommited    []LogEntry
}

// HandleVoteRequest implements APIService.
func (srv *Server) HandleVoteRequest(ctx context.Context, req *VoteRequest) (*VoteResponse, error) {
	var resp VoteResponse

	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.state == FOLLOWER_STATE && req.Term >= srv.Term && srv.Leader == "" {
		srv.election.Reset(electionTimeout())
		resp.Accept = true
		if srv.voteFor == "" {
			srv.voteFor = req.From
			srv.Term = req.Term
		} else if srv.voteFor == req.From && req.Term == srv.Term {
			srv.Leader = req.From
			srv.voteFor = ""
		} else {
			resp.Accept = false
		}
	}
	return &resp, nil
}

func (srv *Server) HandleAppendLogEntry(ctx context.Context, req *AppendLogRequest) (*AppendLogResponse, error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	accept := false
	if srv.state == FOLLOWER_STATE && req.Term == srv.Term && srv.Leader == req.Name {
		accept = true
	}
	srv.election.Reset(electionTimeout())
	// accepts := 0
	// for _, node := range srv.Others {
	// 	res, err := node.SendAppendLogRequest(ctx, &AppendLogRequest{Commited: false})
	// 	if err != nil {
	// 		logger.FromContext(ctx).Errorf("SendAppendLogRequest(node=%v) error: %v\n", node, err)
	// 	}
	// 	if res.Accepted {
	// 		accepts++
	// 	}
	// }
	// if accepts >= (len(srv.Others)+1)/2 {
	// 	for _, node := range srv.Others {
	// 		node.SendAppendLogRequest(ctx, &AppendLogRequest{Commited: true})
	// 	}
	// }

	return &AppendLogResponse{Accept: accept}, nil
}

func raftLeaderLoop(ctx context.Context, me *Server) error {
	heartbeatTick := time.Tick(1 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-heartbeatTick:
			logger.FromContext(ctx).Infof("Now, I(%s) am leader tick tock\n", me.Addr)
			acks := 0
			for _, node := range me.Others {
				_, err := node.SendHealthcheckRequest(ctx, me.Addr, me.Term)
				if err != nil {
					logger.FromContext(ctx).Errorf("SendHealthRequest(node=%v) error: %v\n", node, err)
					continue
				}
				acks++
			}
			if acks == 0 {
				util.BUG(fmt.Sprintf("None of other nodes response my healthcheck request"))
			}
		}
	}
}

func TryRaftLeaderElection(ctx context.Context, me *Server) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-me.election.C:
			log := logger.FromContext(ctx)
			log.Infof("[candidate timeout: %s]: try to election in my term: %d...\n",
				me.Addr, me.Term)

			me.mu.Lock()
			me.state = CANDIDATE_STATE
			me.Term += 1
			me.voteFor = "me"
			me.mu.Unlock()

			var votes int
			for _, node := range me.Others {
				res, err := node.SendVoteRequest(ctx, &VoteRequest{From: me.Addr, Term: me.Term})
				if err != nil {
					log.Errorf("SendVoteRequest(node=%+v) error: %v\n", node, err)
					continue
				}
				if res.Accept {
					votes++
				}
			}
			if votes >= (len(me.Others)+1)/2 {
				me.mu.Lock()
				me.state = LEADER_STATE
				me.mu.Unlock()
				me.election.Stop()
				if err := raftLeaderLoop(ctx, me); err != nil {
					return err
				}
				me.mu.Lock()
				me.state = FOLLOWER_STATE
				me.mu.Unlock()
			}
			log.Infof("Received %d votes, election failed, try again\n", votes)
			me.election.Reset(electionTimeout())
		}
	}
}

func Run(ctx context.Context, srv *Server) error {
	handler := RegisterHttpHandlers(ctx, srv)

	go util.WithContext(ctx, func() error {
		return httputil.StartHttpServer(ctx, srv.Addr, handler)
	})
	return TryRaftLeaderElection(ctx, srv)
}

var options struct {
	addr    string
	members string
}

func parseOptions() {
	flag.StringVar(&options.addr, "addr", "127.0.0.1:3678", "specify the address of the node")
	flag.StringVar(&options.members, "members", "", "specify the members of the cluster")
	flag.Parse()
	if options.addr == "" {
		util.Fatal("addr of the server must not be empty")
	}
	if options.members == "" {
		util.Fatal("members must not be empty")
	}
}

func initClusterMembers() []*Node {
	var nodes []*Node

	for _, memb := range strings.Split(options.members, ",") {
		nodes = append(nodes, &Node{
			Addr: memb,
			Client: &httpClient{
				endpoint: memb, Client: http.DefaultClient,
			}})
	}
	return nodes
}

func main() {
	parseOptions()
	nodes := initClusterMembers()
	server := Server{Addr: options.addr, Others: nodes, Term: 0, election: time.NewTimer(electionTimeout())}
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
