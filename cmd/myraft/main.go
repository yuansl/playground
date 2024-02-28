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

	"github.com/yuansl/playground/util"
)

type Role int

const (
	Follower Role = iota
	Candidate
	Leader
)

type State struct {
	Addr    string
	Role    Role
	Term    int
	Others  []State
	logchan chan *LogEntry
}

func (state *State) HandleAppendLogEntry(ctx context.Context) (*LogEntryAck, error) {
	state.logchan <- &LogEntry{} // TODO:
	return &LogEntryAck{Accepted: true}, nil
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

type VoteRequest struct {
	Term int
	From *Node
}

type VoteResponse struct {
	Term  int
	Voted bool
	Peer  string
}

func SendVoteRequest(addr string, req *VoteRequest) (*VoteResponse, error) {
	return &VoteResponse{Peer: "B", Voted: true}, nil
}

type HealthRequest struct {
	PeerAddr string
	Message  []byte
}

type HealthResponse struct {
	Ack string
}

func SendHealthRequest(req *HealthRequest) (*HealthResponse, error) {
	_ = req.PeerAddr
	return &HealthResponse{Ack: "OKay"}, nil
}

func leaderLoop(me *State) {
	for {
		heartbeatTick := time.Tick(10 * time.Millisecond)
		select {
		case <-heartbeatTick:
			for _, n := range me.Others {
				SendHealthRequest(&HealthRequest{PeerAddr: n.Addr})
			}
		}
	}
}

type LogEntry struct{}

type LogEntryAck struct{ Accepted bool }

type APIService interface {
	HandleAppendLogEntry(ctx context.Context) (*LogEntryAck, error)
}

func RegisterHttpHandlers(ctx context.Context, srv APIService) http.Handler {
	handler := httputil.InitHttpHandlerRegister()

	handler.GET("/v1/appendlog", func(c *httputil.HttpContext) {
		var request struct {
			/* TODO: add missing fields */
		}
		httputil.HandleRequest(c, &request, func(ctx context.Context) (any, error) {
			return srv.HandleAppendLogEntry(ctx)
		})
	})
	handler.GET("/v1/heartbeat", func(ctx *httputil.HttpContext) {
		ctx.JSON(http.StatusOK, &httputil.Response{Result: "PONG"})
	})
	return handler.(http.Handler)
}

func TryElectLeader(me *State) error {
	logch := make(chan *LogEntry)

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
			}
		case <-logch:
		}
	}
}

func Run(ctx context.Context, state *State) error {
	handler := RegisterHttpHandlers(ctx, state)

	go netutil.WithContext(ctx, func() error {
		return httputil.StartHttpServer(ctx, state.Addr, handler)
	})
	return TryElectLeader(state)
}

func main() {
	me := State{Addr: "A", Others: []State{{Addr: "B"}, {Addr: "C"}}, Term: 0}
	ctx := util.InitSignalHandler(context.TODO())

	Run(ctx, &me)
}
