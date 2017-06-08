package services

import (
	"errors"

	"golang.org/x/net/context"

	"fmt"

	"github.com/unerror/waffy/pkg/repository"
	"github.com/unerror/waffy/pkg/services/protos/nodes"
)

// Node is the JoinService implementation for a Node
type Node struct {
	baseHandler
}

// Join joins a new Node to the RPC and Raft consensus
func (n *Node) Join(ctx context.Context, req *nodes.JoinRequest) (*nodes.JoinResponse, error) {
	if _, err := repository.FindNodeByHostname(n.s, req.Hostname); err != nil {
		return nil, errors.New("hostname already exists")
	}

	node := &nodes.Node{
		Hostname: req.Hostname,
		Leader:   false, // not leader by default until election
	}
	if err := repository.CreateNode(n.s, node); err != nil {
		return nil, errors.New("unable to create new node")
	}

	err := n.s.Join(req.Hostname)
	if err != nil {
		return nil, fmt.Errorf("unable to join raft node: %s", err)
	}

	return &nodes.JoinResponse{
		Hostname: req.Hostname,
	}, nil
}

// Leave leaves a Node from the RPC and Raft consensus
func (n *Node) Leave(ctx context.Context, req *nodes.LeaveRequest) (*nodes.LeaveResponse, error) {
	if err := repository.DeleteNodeByHostname(n.s, req.Hostname); err != nil {
		return nil, errors.New("error deleting node in store")
	}
	if err := n.s.Leave(req.Hostname); err != nil {
		return nil, fmt.Errorf("unable to leave raft: %s", err)
	}

	return &nodes.LeaveResponse{
		Hostname: req.Hostname,
	}, nil
}
