package services

import (
	"golang.org/x/net/context"

	"github.com/unerror/waffy/pkg/data"
	"github.com/unerror/waffy/pkg/services/protos/nodes"
	"github.com/unerror/waffy/pkg/services/protos/certificates"
)

type Node struct {
	hostname string
	certificate certificates.Certificate
	raft *data.Raft
}

func (n *Node) Join(ctx context.Context, req nodes.JoinRequest) (*nodes.JoinResponse, error) {
	return &nodes.JoinResponse{
		Error: n.raft.Join(req.Url).Error(),
	}
}

func (n *Node) Leave(ctx context.Context, req nodes.LeaveRequest) (*nodes.LeaveResponse, error) {
	return &nodes.LeaveResponse{
		Error: n.raft.Leave(req.Url).Error(),
	}
}
