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
	"time"

	"github.com/qbox/net-deftones/logger"
	netutil "github.com/qbox/net-deftones/util"
	"github.com/qbox/net-deftones/util/httputil"
)

type State int

const (
	Follower State = iota
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
	Peer  *Node
}

type AppendLogRequest struct {
	Commited bool
	Value    LogEntry
	Term     int
}

type AppendLogResponse struct {
	Accepted bool
}

type RaftService interface {
	HandleAppendLogEntry(ctx context.Context, req *AppendLogRequest) (*AppendLogResponse, error)
	HandleVoteRequest(ctx context.Context, req *VoteRequest) (*VoteResponse, error)
	SendHealthRequest(req *AppendLogRequest) (*AppendLogResponse, error)
	SendAppendLogRequest(req *AppendLogRequest) (*AppendLogResponse, error)
}

type LogEntry struct {
	Term  int
	Value any
}

type Server struct {
	Addr          string
	Leader        string
	state         State
	Term          int
	Others        []*Node
	appendlogChan chan *AppendLogRequest
	voteFor       string
	commited      []LogEntry
	uncommited    []LogEntry
}

// HandleVoteRequest implements APIService.
func (*Server) HandleVoteRequest(ctx context.Context, req *VoteRequest) (*VoteResponse, error) {
	panic("unimplemented")
}

func (srv *Server) HandleAppendLogEntry(ctx context.Context, req *AppendLogRequest) (*AppendLogResponse, error) {
	if srv.state != Follower {
		return nil, fmt.Errorf("Invalid state")
	}
	srv.appendlogChan <- req
	return &AppendLogResponse{Accepted: true}, nil
}

func (*Server) HandleHealthRequest(req *AppendLogRequest) (*AppendLogResponse, error) {
	return &AppendLogResponse{Accepted: true}, nil
}

func (*Server) HandleAppendLogPrepare(req *AppendLogRequest) (*AppendLogResponse, error) {
	return &AppendLogResponse{Accepted: true}, nil
}

func (*Server) HandleAppendLogCommit(req *AppendLogRequest) (*AppendLogResponse, error) {
	return &AppendLogResponse{Accepted: true}, nil
}

type Node struct {
	Addr string
	Name string
}

type VoteStatus int

const (
	VoteProgressing VoteStatus = iota
	VoteWin
)

func (*Node) SendVoteRequest(addr string, req *VoteRequest) (*VoteResponse, error) {
	return &VoteResponse{Peer: &Node{Name: "B"}, Voted: true}, nil
}

func (node *Node) SendHealthRequest(term int) (*AppendLogResponse, error) {
	return node.SendAppendLogPrepare(&AppendLogRequest{Term: term})
}

func (*Node) SendAppendLogPrepare(req *AppendLogRequest) (*AppendLogResponse, error) {
	return &AppendLogResponse{Accepted: true}, nil
}

func (*Node) SendAppendLogCommit(req *AppendLogRequest) (*AppendLogResponse, error) {
	return &AppendLogResponse{}, nil
}

func raftLeaderLoop(ctx context.Context, me *Server) error {
	heartbeatTick := time.Tick(10 * time.Millisecond)
	log := logger.FromContext(ctx)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-heartbeatTick:
			acks := 0
			for _, node := range me.Others {
				_, err := node.SendHealthRequest(me.Term)
				if err != nil {
					log.Errorf("SendHealthRequest(node=%v) error: %v\n", node, err)
					continue
				}
				acks += 1
			}
			if acks == 0 {
				return fmt.Errorf("None of others response my healthcheck request")
			}

		case <-me.appendlogChan: // new request from client
			accepts := 0
			for _, node := range me.Others {
				res, err := node.SendAppendLogPrepare(&AppendLogRequest{})
				if err != nil {
					log.Errorf("SendAppendLogPrepare(node=%v) error: %v\n", node, err)
				}
				if res.Accepted {
					accepts++
				}
			}
			if accepts >= (len(me.Others)+1)/2 {
				for _, node := range me.Others {
					node.SendAppendLogCommit(&AppendLogRequest{Commited: true})
				}
			}
		}
	}
}

func TryElectRaftLeader(ctx context.Context, me *Server) error {
	votechan := make(chan *VoteRequest)
	electionTimeout := func() time.Duration { return time.Duration(rand.Int()%151+150) * time.Millisecond }
	electionTimer := time.NewTimer(electionTimeout())

	for {
		select {
		case <-electionTimer.C:
			electionTimer.Reset(electionTimeout())
			me.state = Candidate
			me.Term += 1

			var votes int
			node0 := Node{Addr: me.Addr, Name: "A"}
			for _, node := range me.Others {
				res, err := node.SendVoteRequest(node.Addr, &VoteRequest{From: &node0, Term: me.Term})
				if err != nil {
					fmt.Printf("SendVoteRequest(node=%+v) error: %v\n", node, err)
					continue
				}
				if res.Voted {
					votes++
				}
			}
			if votes >= (len(me.Others)+1)/2 {
				me.state = Leader
				raftLeaderLoop(ctx, me)
				me.state = Follower
			}
		case voteReq := <-votechan:
			if me.state == Follower && me.voteFor == "" {
				me.voteFor = voteReq.From.Name
				me.Term = voteReq.Term
				// Accepted = true
			} else {
				// Accepted = false
			}
		}
	}
}

func Run(ctx context.Context, state *Server) error {
	handler := RegisterHttpHandlers(ctx, state)

	go netutil.WithContext(ctx, func() error {
		return httputil.StartHttpServer(ctx, state.Addr, handler)
	})
	return TryElectRaftLeader(ctx, state)
}

func main() {
	// server := Server{Addr: "A", Others: []Server{{Addr: "B"}, {Addr: "C"}}, Term: 0}

	// ctx := netutil.InitSignalHandler(context.TODO())
	// err := Run(ctx, &server)
	// if err != nil {
	// 	switch {
	// 	case errors.Is(err, context.Canceled):
	// 		err = context.Cause(ctx)
	// 		fallthrough
	// 	default:
	// 		util.Fatal(err)
	// 	}
	// }
}
