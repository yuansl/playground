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
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"

	netutil "github.com/qbox/net-deftones/util"
	"github.com/qbox/net-deftones/util/httputil"
)

type Role int

const (
	Follower Role = iota
	Candidate
	Leader
)

type VoteRequest struct {
	Term int
	From *Node
}

type VoteResponse struct {
	Term  int
	Voted bool
	Peer  string
}

type AppendLogRequest struct {
	PeerAddr string
	Commited bool
	Value    LogEntry
}

type AppendLogResponse struct {
	Accepted bool
}

type RaftService interface {
	HandleAppendLogEntry(ctx context.Context, req *AppendLogRequest) (*AppendLogResponse, error)
	HandleVoteRequest(ctx context.Context, req *VoteRequest) (*VoteResponse, error)
	SendHealthRequest(req *AppendLogRequest) (*AppendLogResponse, error)
}

type LogEntry struct {
	Term  int
	Value any
}

type State struct {
	Addr          string
	Leader        string
	Role          Role
	Term          int
	Others        []State
	appendlogChan chan *AppendLogRequest
	voteFor       string
	commited      []LogEntry
	uncommited    []LogEntry
}

// HandleVoteRequest implements APIService.
func (state *State) HandleVoteRequest(ctx context.Context, req *VoteRequest) (*VoteResponse, error) {
	panic("unimplemented")
}

func (state *State) HandleAppendLogEntry(ctx context.Context, req *AppendLogRequest) (*AppendLogResponse, error) {
	if state.Role != Follower || state.Leader != req.PeerAddr {
		return nil, fmt.Errorf("Invalid state")
	}
	state.appendlogChan <- req
	return &AppendLogResponse{Accepted: true}, nil
}

var _ RaftService = (*State)(nil)

type Node struct {
	Addr string
	Name string
}

type VoteStatus int

const (
	VoteProgressing VoteStatus = iota
	VoteWin
)

func SendVoteRequest(addr string, req *VoteRequest) (*VoteResponse, error) {
	return &VoteResponse{Peer: "B", Voted: true}, nil
}

func (*State) SendHealthRequest(req *AppendLogRequest) (*AppendLogResponse, error) {
	return &AppendLogResponse{Accepted: true}, nil
}

func (*State) SendAppendLogPrepare(req *AppendLogRequest) (*AppendLogResponse, error) {
	return &AppendLogResponse{Accepted: true}, nil
}

func leaderLoop(ctx context.Context, me *State) error {
	for {
		heartbeatTick := time.Tick(10 * time.Millisecond)
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-heartbeatTick:
			for _, n := range me.Others {
				me.SendHealthRequest(&AppendLogRequest{
					Value: LogEntry{Term: me.Term, Value: "PING"},
				})
			}

		case <-me.appendlogChan:
			for _, n := range me.Others {
				me.SendAppendLogPrepare(&AppendLogRequest{})
			}
		}
	}
}

func RegisterHttpHandlers(ctx context.Context, srv RaftService) http.Handler {
	handler := httputil.InitHttpHandlerRegister()

	handler.GET("/v1/appendlog", func(c httputil.HttpContext) {
		var request AppendLogRequest
		httputil.HandleRequest(c, &request, func(ctx context.Context) (any, error) {
			return srv.HandleAppendLogEntry(ctx, &request)
		})
	})
	handler.GET("/v1/vote", func(c httputil.HttpContext) {
		var request VoteRequest
		httputil.HandleRequest(c, &request, func(ctx context.Context) (any, error) {
			return srv.HandleVoteRequest(ctx, &request)
		})
	})

	return handler.(http.Handler)
}

func TryElectLeader(ctx context.Context, me *State) error {
	votechan := make(chan *VoteRequest)

	for {
		electionTimer := time.NewTimer(time.Duration(rand.Int()%151+150) * time.Millisecond)

		select {
		case <-electionTimer.C:
			me.Role = Candidate
			me.Term += 1

			var votes int
			node0 := Node{Addr: me.Addr, Name: "A"}
			for _, node := range me.Others {
				res, err := SendVoteRequest(node.Addr, &VoteRequest{From: &node0, Term: me.Term})
				if err != nil {
					fmt.Printf("SendVoteRequest(node=%+v) error: %v\n", node, err)
					continue
				}
				if res.Voted {
					votes++
				}
			}
			if votes >= len(me.Others)/2 {
				me.Role = Leader
				leaderLoop(ctx, me)
			}
		case r := <-votechan:
			if me.Role == Candidate || (me.Term == r.Term && me.voteFor != "") {
				// Accepted = false
			} else {
				// Accepted = true
			}
		}
	}
}

func Run(ctx context.Context, state *State) error {
	handler := RegisterHttpHandlers(ctx, state)

	go netutil.WithContext(ctx, func() error {
		return httputil.StartHttpServer(ctx, state.Addr, handler)
	})
	return TryElectLeader(ctx, state)
}

func main() {
	me := State{Addr: "A", Others: []State{{Addr: "B"}, {Addr: "C"}}, Term: 0}
	ctx := netutil.InitSignalHandler(context.TODO())

	Run(ctx, &me)
}
