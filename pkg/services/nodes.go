package services

import (
	"errors"
	"crypto/x509"

	"golang.org/x/net/context"

	"github.com/unerror/waffy/pkg/config"
	"github.com/unerror/waffy/pkg/data"
	"github.com/unerror/waffy/pkg/services/protos/nodes"
)

type Node struct {
	hostname string
	certificate *x509.Certificate
	node data.Consensus
}

func NewRaftNode() (*Node,error) {

	n := new(Node)

	conf, err := config.Load()
	if err != nil {
		return n, err
	}

	certPath := conf.CertPath
	cert, err := config.LoadCert(certPath)
	if err != nil {
		return n, err
	}

	store, err := data.NewDB(conf.DBPath)
	if err != nil {
		return n, err
	}
	raftDir := conf.RaftDIR
	raftListen := conf.RaftListen

	n.hostname = conf.RPCName
	n.certificate = cert

	n.node, err = data.NewRaft(raftDir,raftListen,store);
	if err != nil{
		return &Node{}, errors.New("Unable to connect to Raft")
	}

	return n, nil

}

func (n *Node) Join(ctx context.Context, req nodes.JoinRequest) (*nodes.JoinResponse, error) {
	err := n.node.Join(req.Url)
	return &nodes.JoinResponse{
		Hostname: n.hostname,
		Signature: n.certificate.Signature,
		Error: err.Error(),
	}, err
}

func (n *Node) Leave(ctx context.Context, req nodes.LeaveRequest) (*nodes.LeaveResponse, error) {
	err := n.node.Leave(req.Url)
	return &nodes.LeaveResponse{
		Hostname: n.hostname,
		Signature: n.certificate.Signature,
		Error: err.Error(),
	}, err
}
