package services

import (
	"errors"

	"golang.org/x/net/context"

	"fmt"

	"github.com/unerror/waffy/pkg/repository"
	"github.com/unerror/waffy/pkg/services/protos/nodes"
)

type Node struct {
	baseHandler
}

func (n *Node) Join(ctx context.Context, req *nodes.JoinRequest) (*nodes.JoinResponse, error) {
	if _, err := repository.FindNodeByHostname(n.s, req.Url); err != nil {
		return nil, errors.New("hostname already exists")
	}

	node := &nodes.Node{
		Hostname: req.Hostname,
		Leader:   false, // not leader by default until election
	}
	if err := repository.CreateNode(n.s, node); err != nil {
		return nil, errors.New("unable to create new node")
	}

	err := n.s.Join(req.Url)
	if err != nil {
		return nil, fmt.Errorf("unable to join raft node: %s", err)
	}

	return &nodes.JoinResponse{
		Hostname: req.Url,
	}, nil
}

func (n *Node) Leave(ctx context.Context, req *nodes.LeaveRequest) (*nodes.LeaveResponse, error) {
	if err := repository.DeleteNodeByHostname(n.s, req.Url); err != nil {
		return nil, errors.New("error deleting node in store")
	}
	if err := n.s.Leave(req.Url); err != nil {
		return nil, fmt.Errorf("unable to leave raft: %s", err)
	}

	return &nodes.LeaveResponse{
		Hostname: req.Url,
	}, nil
}
