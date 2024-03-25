package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/qbox/net-deftones/logger"
	"github.com/qbox/net-deftones/util"
)

type RaftServer struct {
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
	uncommited    *LogEntry
}

func (srv *RaftServer) HandleGetlogRequest(ctx context.Context, req *GetlogRequest) (*GetlogResponse, error) {
	srv.mu.RLock()
	defer srv.mu.RUnlock()

	var logent LogEntry

	for _, it := range srv.commited {
		if it.Key == req.Key {
			logent = it
			break
		}
	}
	return &GetlogResponse{Value: logent.Value}, nil
}

func (srv *RaftServer) HandleAddlogRequest(ctx context.Context, req *AddlogRequest) (*AddlogResponse, error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	// prepare-commit
	srv.uncommited = &LogEntry{Key: req.Key, Value: req.Value, Term: srv.Term}

	accept := false
	accepts := 0
	for _, node := range srv.Others {
		res, err := node.SendAppendLogRequest(ctx, &AppendLogRequest{Value: srv.uncommited, Term: srv.Term, Name: srv.Addr, Commited: false})
		if err != nil {
			logger.FromContext(ctx).Warnf("SendAppendLogRequest error: %v\n", err)
			continue
		}
		if res.Accept {
			accepts += 1
		}
	}
	// commit
	if accepts >= (len(srv.Others)+1)/2 {
		for _, node := range srv.Others {
			_, err := node.SendAppendLogRequest(ctx, &AppendLogRequest{
				Value: srv.uncommited, Term: srv.Term, Name: srv.Addr, Commited: true})
			if err != nil {
				logger.FromContext(ctx).Warnf("SendAppendLogRequest error: %v\n", err)
				continue
			}
		}
		srv.commited = append(srv.commited, *srv.uncommited)
		srv.uncommited = nil
		accept = true
	}

	return &AddlogResponse{Accept: accept}, nil
}

// HandleVoteRequest implements APIService.
func (srv *RaftServer) HandleVoteRequest(ctx context.Context, req *VoteRequest) (*VoteResponse, error) {
	var resp VoteResponse

	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.state == FOLLOWER_STATE && req.Term >= srv.Term && srv.Leader == "" {
		srv.election.Reset(electionTimeout())
		if srv.voteFor == "" {
			resp.Accept = true
			srv.voteFor = req.From
			srv.Term = req.Term
		} else {
			resp.Accept = false
		}
	}
	return &resp, nil
}

func (srv *RaftServer) HandleAppendLogEntry(ctx context.Context, req *AppendLogRequest) (*AppendLogResponse, error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	accept := false
	if srv.state != LEADER_STATE && req.Term >= srv.Term {
		if srv.voteFor == req.Name && req.Term == srv.Term {
			srv.Leader = req.Name
			srv.voteFor = ""
		} else if srv.Leader == "" {
			srv.Leader = req.Name
			srv.Term = req.Term
		}
		if req.Name == srv.Leader {
			accept = true
		}
	}
	if accept {
		if req.Value.Value != "PING" {
			if !req.Commited {
				srv.uncommited = req.Value
			} else {
				srv.commited = append(srv.commited, *srv.uncommited)
				srv.uncommited = nil
			}
		}
	}
	srv.election.Reset(electionTimeout())

	return &AppendLogResponse{Accept: accept}, nil
}

func (me *RaftServer) raftLeaderLoop(ctx context.Context) error {
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

func (me *RaftServer) TryRaftLeaderElection(ctx context.Context) error {
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
				if err := me.raftLeaderLoop(ctx); err != nil {
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
