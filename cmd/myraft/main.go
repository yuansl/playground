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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
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

type Request struct {
	Path        string
	Method      string
	Body        []byte
	ContentType string
}

type Response = httputil.Response

type Client interface {
	Send(ctx context.Context, _ *Request) (*Response, error)
}

type Node struct {
	Client
	Addr string
	Name string
}

func (node *Node) SendVoteRequest(addr string, req *VoteRequest) (*VoteResponse, error) {
	data, _ := json.Marshal(req)
	resp, err := node.Send(context.Background(), &Request{Path: "/v1/vote", Method: http.MethodPost, Body: data, ContentType: "application/json"})
	if err != nil {
		return nil, err
	}
	_ = resp
	return &VoteResponse{Peer: &Node{Name: "B"}, Voted: true}, nil
}

func (node *Node) SendHealthcheckRequest(term int) (*AppendLogResponse, error) {
	return node.SendAppendLogRequest(&AppendLogRequest{Term: term, Commited: true})
}

func (node *Node) SendAppendLogRequest(req *AppendLogRequest) (*AppendLogResponse, error) {
	data, _ := json.Marshal(req)
	resp, err := node.Send(context.Background(), &Request{Path: "/v1/appendlog", Method: http.MethodPost, Body: data, ContentType: "application/json"})
	if err != nil {
		return nil, err
	}
	return &AppendLogResponse{Accepted: resp.Result.(map[string]any)["Accepted"].(bool)}, nil
}

type client struct {
	*http.Client
	endpoint string
}

// Do implements Client.
func (c *client) Send(ctx context.Context, req *Request) (*Response, error) {
	var body bytes.Buffer
	if len(req.Body) > 0 {
		body.Write(req.Body)
	}
	URL := c.endpoint + req.Path
	hreq, err := http.NewRequestWithContext(ctx, req.Method, URL, &body)
	if err != nil {
		return nil, err
	}
	hres, err := c.Client.Do(hreq)
	if err != nil {
		return nil, err
	}
	var res Response
	if err = json.NewDecoder(hres.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

var _ Client = (*client)(nil)

type VoteRequest struct {
	Term int
	From string
}

type VoteResponse struct {
	Term  int
	Voted bool
	Peer  *Node
}

type AppendLogRequest struct {
	Term     int
	Commited bool
	Value    *LogEntry
}

type AppendLogResponse struct {
	Accepted bool
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
	state         State
	Term          int
	Others        []*Node
	election      *time.Timer
	healthcheck   *time.Ticker
	appendlogChan chan *AppendLogRequest
	voteFor       string
	commited      []LogEntry
	uncommited    []LogEntry
}

// HandleVoteRequest implements APIService.
func (srv *Server) HandleVoteRequest(ctx context.Context, req *VoteRequest) (*VoteResponse, error) {
	var resp VoteResponse

	if srv.state == FOLLOWER_STATE && req.Term >= srv.Term && srv.Leader == "" {
		if srv.voteFor == "" {
			resp.Term = req.Term
			resp.Voted = true
			srv.Term = req.Term
			srv.voteFor = req.From
		} else if srv.voteFor == req.From {
			srv.voteFor = ""
			srv.Leader = req.From
		}
		srv.election.Reset(electionTimeout())
	}
	return &resp, nil
}

func (srv *Server) HandleAppendLogEntry(ctx context.Context, req *AppendLogRequest) (*AppendLogResponse, error) {
	if srv.state != FOLLOWER_STATE {
		return nil, ErrInvalidState
	}
	accepts := 0
	for _, node := range srv.Others {
		res, err := node.SendAppendLogRequest(&AppendLogRequest{Commited: false})
		if err != nil {
			logger.FromContext(ctx).Errorf("SendAppendLogPrepare(node=%v) error: %v\n", node, err)
		}
		if res.Accepted {
			accepts++
		}
	}
	if accepts >= (len(srv.Others)+1)/2 {
		for _, node := range srv.Others {
			node.SendAppendLogRequest(&AppendLogRequest{Commited: true})
		}
	}

	return &AppendLogResponse{Accepted: true}, nil
}

func raftLeaderLoop(ctx context.Context, me *Server) error {
	heartbeatTick := time.Tick(10 * time.Millisecond)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-heartbeatTick:
			logger.FromContext(ctx).Infof("tick tock\n")
			acks := 0
			for _, node := range me.Others {
				_, err := node.SendHealthcheckRequest(me.Term)
				if err != nil {
					logger.FromContext(ctx).Errorf("SendHealthRequest(node=%v) error: %v\n", node, err)
					continue
				}
				acks++
			}
			if acks == 0 {
				return fmt.Errorf("None of others response my healthcheck request")
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
			me.state = CANDIDATE_STATE
			me.Term += 1
			me.voteFor = "me"

			logger.FromContext(ctx).Infof("Try to election in my term: %d...\n", me.Term)

			var votes int
			for _, node := range me.Others {
				res, err := node.SendVoteRequest(node.Addr, &VoteRequest{From: me.Addr, Term: me.Term})
				if err != nil {
					fmt.Printf("SendVoteRequest(node=%+v) error: %v\n", node, err)
					continue
				}
				if res.Voted {
					votes++
				}
			}
			if votes >= (len(me.Others)+1)/2 {
				me.state = LEADER_STATE
				me.election.Stop()
				if err := raftLeaderLoop(ctx, me); err != nil {
					return err
				}
				me.state = FOLLOWER_STATE
			}
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
}

func main() {
	parseOptions()
	var nodes []*Node
	server := Server{Addr: options.addr, Others: nodes, Term: 0, election: time.NewTimer(electionTimeout())}
	ctx := util.InitSignalHandler(context.TODO())

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
