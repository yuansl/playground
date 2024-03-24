package main

import (
	"context"
	"encoding/json"
	"net/http"
)

type Node struct {
	Client
	Addr string
	Name string
}

func (node *Node) SendVoteRequest(ctx context.Context, req *VoteRequest) (*VoteResponse, error) {
	data, _ := json.Marshal(req)
	resp, err := node.Send(ctx, &Request{Path: "/v1/vote", Method: http.MethodPost, Body: data, ContentType: "application/json"})
	if err != nil {
		return nil, err
	}
	return &VoteResponse{Accept: resp.Result.(map[string]any)["accept"].(bool)}, nil
}

func (node *Node) SendHealthcheckRequest(ctx context.Context, from string, term int) (*AppendLogResponse, error) {
	return node.SendAppendLogRequest(ctx, &AppendLogRequest{Name: from, Term: term, Commited: true, Value: &LogEntry{Value: "PING"}})
}

func (node *Node) SendAppendLogRequest(ctx context.Context, req *AppendLogRequest) (*AppendLogResponse, error) {
	data, _ := json.Marshal(req)
	resp, err := node.Send(ctx, &Request{Path: "/v1/appendlog", Method: http.MethodPost, Body: data, ContentType: "application/json"})
	if err != nil {
		return nil, err
	}
	return &AppendLogResponse{Accept: resp.Result.(map[string]any)["accept"].(bool)}, nil

}
