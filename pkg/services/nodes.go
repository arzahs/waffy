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
func (n *Node) Join(ctx context.Context, req *nodes.NodeRequest) (*nodes.JoinResponse, error) {
	if n, err := repository.FindNodeByHostname(n.s, req.Node.Hostname); n != nil && err != nil {
		return nil, fmt.Errorf("hostname already exists: %s", err)
	}

	node := &nodes.Node{
		Hostname: req.Node.Hostname,
		Leader:   false, // not leader by default until election
	}
	if err := repository.SetNode(n.s, node); err != nil {
		return nil, fmt.Errorf("unable to create new node: %s", err)
	}

	err := n.s.Join(req.Node.RaftAddress)
	if err != nil {
		return nil, fmt.Errorf("unable to join raft node: %s", err)
	}

	return &nodes.JoinResponse{
		Hostname: req.Node.Hostname,
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

// UpdateNode updates information about a Node
func (n *Node) UpdateNode(ctx context.Context, req *nodes.NodeRequest) (*nodes.UpdateNodeResponse, error) {
	res := &nodes.UpdateNodeResponse{
		Node: req.Node,
		New:  false,
	}
	existing, err := repository.FindNodeByHostname(n.s, req.Node.Hostname)
	if err != nil && existing == nil {
		return nil, fmt.Errorf("unable to save node: %s", err)
	}

	if existing == nil {
		res.New = true
	}

	err = repository.SetNode(n.s, req.Node)
	if err != nil {
		return nil, err
	}

	return res, nil
}
