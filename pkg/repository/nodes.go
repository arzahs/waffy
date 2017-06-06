package repository

import (
	"github.com/unerror/waffy/pkg/data"
	"github.com/unerror/waffy/pkg/services/protos/nodes"
)

const (
	// NodesBucket defines the Store bucket used for Nodes
	NodesBucket = "nodes"
)

// CreateNode creates a Node in the data Store
func CreateNode(d data.Store, n *nodes.Node) error {
	b, err := d.Bucket(NodesBucket)
	if err != nil {
		return err
	}

	return Create(b, []byte(n.Hostname), n)
}

// FindNodeByHostname finds a Node in the data Store by it's hostname
func FindNodeByHostname(d data.Store, hostname string) (*nodes.Node, error) {
	b, err := d.Bucket(NodesBucket)
	if err != nil {
		return nil, err
	}

	n := nodes.Node{}
	err = Seek(b, []byte(hostname), &n)
	if err != nil {
		return nil, err
	}

	return &n, nil
}

// DeleteNodeByHostname removes a given Node by hostname
func DeleteNodeByHostname(d data.Store, hostname string) error {
	b, err := d.Bucket(NodesBucket)
	if err != nil {
		return err
	}

	return DeleteByKey(b, []byte(hostname))
}
