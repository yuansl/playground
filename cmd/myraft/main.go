package main

import "time"

type Role int

const (
	Follower Role = iota
	Candidate
	Leader
)

type RaftServer struct {
	id       string
	leaderid string
	role     Role
	term     int
	voteq    chan *VoteRequest
	logq     chan *AppendEntryRequest
	votedFor string
	nodes    []*Node
	options  *serverOptions
}

type serverOptions struct {
	electionTimeout time.Duration
	heartbeat       time.Duration
}

const (
	Reject = iota
	Aggrement
)

type VoteRequest struct {
	Term     int
	LeaderId string
}

type VoteResponse struct {
	Ack int
}

func SendVoteRequest(node *Node, req *VoteRequest) (*VoteResponse, error) {
	return &VoteResponse{Ack: Aggrement}, nil
}

func ReplyVoteRequest() {}

func SendAppendEntryRequest(*Node, *AppendEntryRequest) {

}

const (
	Heartbeat = iota
	AppendLogEntry
)

type AppendEntryRequest struct {
	Type         int
	LeaderId     string
	Term         int
	Message      []byte
	LastLogTerm  int
	LastLogIndex int
}

type Node struct {
	Id           string
	Term         int
	PrevTerm     int
	PrevLogIndex int
}

type Log struct {
	Index   int
	Term    int
	Message []byte
}

func (server *RaftServer) LeaderLoop() {
	heartbeatTicker := time.NewTicker(server.options.heartbeat)

	for {
		select {
		case <-server.logq:
			for _, node := range server.nodes {
				SendAppendEntryRequest(node, &AppendEntryRequest{
					Type: AppendLogEntry,
				})
			}

		case <-heartbeatTicker.C:
			for _, node := range server.nodes {
				SendAppendEntryRequest(node, &AppendEntryRequest{
					Type: Heartbeat,
				})
			}
		}
	}
}

func (server *RaftServer) FollowerLoop() {
	electionTicker := time.NewTicker(server.options.electionTimeout)

	for {
		select {
		case <-electionTicker.C:
			server.role = Candidate

			votes := 1
			for _, node := range server.nodes {
				r, err := SendVoteRequest(node, &VoteRequest{Term: server.term + 1, LeaderId: server.leaderid})
				if err != nil {
					continue
				}
				if r.Ack == Aggrement {
					votes++
				}
			}
			if votes >= len(server.nodes)/2 {
				server.role = Leader
				server.LeaderLoop()
			}

		case voteReq := <-server.voteq:
			electionTicker.Reset(server.options.electionTimeout)
			server.role = Follower

			if voteReq.Term > server.term && server.votedFor == "" {
				server.votedFor = voteReq.LeaderId
			}
			ReplyVoteRequest()

		case logReq := <-server.logq:
			if logReq.Term == server.term && logReq.LeaderId == server.votedFor {
				if server.leaderid == "" {
					server.leaderid = logReq.LeaderId
				}
			}
		}
	}
}

func (server *RaftServer) start() {
	server.FollowerLoop()
}

func main() {
	server := RaftServer{
		options: &serverOptions{},
		voteq:   make(chan *VoteRequest),
	}

	server.start()
}
