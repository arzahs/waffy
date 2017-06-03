package data

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
)

const (
	maxPoolSize = 3
	timeout     = 10 * time.Second
	retain      = 3
)

const (
	opDeleteBucket = iota
	opGet
	opSet
	opSeek
	opList
	opDelete
)

type command struct {
	Op         int
	Key        []byte
	BucketPath string
	Value      []byte
}

// Raft represents a consensus store, which is managed by a Leader and distributed to Nodes. The
// Raft Bucket implements Strong consensus to ensure data reads are consistent across the cluster
type Raft struct {
	s    Store
	r    *raft.Raft
	path string

	l *sync.Mutex
}

func NewRaft(raftDir, raftListen string, d Bucket) (Consensus, error) {
	raftConfig := raft.DefaultConfig()

	addr, err := net.ResolveTCPAddr("tcp", raftListen)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(raftListen, addr, maxPoolSize, timeout, os.Stderr)
	if err != nil {
		return nil, err
	}

	raftStore := raft.NewJSONPeers(raftDir, transport)
	peers, err := raftStore.Peers()
	if err != nil {
		return nil, err
	}

	if len(peers) < 1 {
		raftConfig.DisableBootstrapAfterElect = true
		raftConfig.EnableSingleNode = true
	}

	snapshots, err := raft.NewFileSnapshotStore(raftDir, retain, os.Stderr)
	if err != nil {
		return nil, err
	}

	fileLog := filepath.Join(raftDir, "raft.db")
	logs, err := raftboltdb.NewBoltStore(fileLog)

	r := &Raft{
		s:    d,
		path: "/",
	}
	r.r, err = raft.NewRaft(raftConfig, (*fsm)(r), logs, logs, snapshots, raftStore, transport)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Store returns a new Store Bucket (that implements Consensus as well)
func (s *Raft) Bucket(name string) (Bucket, error) {
	return &Raft{
		s:    s.s,
		r:    s.r,
		path: fmt.Sprintf("%s%s/", s.path, name),
	}, nil
}

// DeleteBucket removes a bucket from a store
func (s *Raft) DeleteBucket(name string) error {
	f, err := s.applyCmd(&command{
		Op:         opDeleteBucket,
		Key:        []byte(name),
		BucketPath: s.path,
	})
	if err != nil {
		return err
	}

	return f.error
}

func (s *Raft) Close() error {
	if err := s.s.Close(); err != nil {
		return err
	}

	f := s.r.Shutdown()

	return f.Error()
}

// Get retrieves a strongly consistent key value from the Bucket
func (s *Raft) Get(k []byte) ([]byte, error) {
	f, err := s.applyCmd(&command{
		Op:         opGet,
		Key:        k,
		BucketPath: s.path,
	})
	if err != nil {
		return nil, err
	}

	return f.node.Value, f.error
}

// GetWeak returns an weakly consistent value in the current store
func (s Raft) GetWeak(k []byte) ([]byte, error) {
	s.l.Lock()
	defer s.l.Unlock()

	bucket, err := s.bucket(s.path)
	if err != nil {
		return nil, err
	}

	return bucket.Get(k)
}

// Set sets a Node n to the consensus store
func (s *Raft) Set(n Node) error {
	f, err := s.applyCmd(&command{
		Op:         opSet,
		Key:        n.Key,
		Value:      n.Value,
		BucketPath: s.path,
	})
	if err != nil {
		return err
	}

	return f.error
}

// Delete deletes a Node n from the consensus store
func (s *Raft) Delete(n Node) error {
	f, err := s.applyCmd(&command{
		Op:         opDelete,
		Key:        n.Key,
		BucketPath: s.path,
	})
	if err != nil {
		return err
	}

	return f.error
}

// List returns the Nodes stored in the Bucket
func (s *Raft) List() ([]Node, error) {
	f, err := s.applyCmd(&command{
		Op:         opList,
		BucketPath: s.path,
	})
	if err != nil {
		return nil, err
	}

	return f.nodes, f.error
}

// ListWeak returns a weakly consistent List of Nodes in the current store
func (s *Raft) ListWeak() ([]Node, error) {
	s.l.Lock()
	defer s.l.Unlock()

	b, err := s.bucket(s.path)
	if err != nil {
		return nil, err
	}
	return b.List()
}

// Seek fins a value in the Bucket by key k
func (s *Raft) Seek(k []byte) ([]byte, error) {
	f, err := s.applyCmd(&command{
		Op:         opSet,
		Key:        k,
		BucketPath: s.path,
	})
	if err != nil {
		return nil, err
	}

	return f.node.Value, f.error
}

// SeekWeak finds a weakly consistent value in the Bucket by key k
func (s *Raft) SeekWeak(k []byte) ([]byte, error) {
	s.l.Lock()
	defer s.l.Unlock()

	b, err := s.bucket(s.path)
	if err != nil {
		return nil, err
	}

	return b.Seek(k)
}

// Join joins another Node to this consensus
func (s *Raft) Join(addr string) error {
	if s.r.State() != raft.Leader {
		return fmt.Errorf("invalid join request for addr %s on non-leader", addr)
	}

	return s.r.AddPeer(addr).Error()
}

// Leave leaves a Raft node from this consensus
func (s *Raft) Leave(addr string) error {
	if s.r.State() != raft.Leader {
		return fmt.Errorf("invalid leave request for addr %s on non-leader", addr)
	}

	return s.r.RemovePeer(addr).Error()
}

// Bucket returns the leak bucket for the given path. Paths are stored as
// slash-separated values, similar to a UNIX file system
func (s *Raft) bucket(path string) (Bucket, error) {
	paths := strings.Split(path, "/")

	var b Bucket
	for _, bName := range paths {
		var err error
		b, err = s.s.Bucket(bName)
		if err != nil {
			return nil, err
		}
	}

	return b, nil
}

// applyCmd applies a command to the Raft Log, and returns the waited for *fsmResponse (or error)
// related
func (s *Raft) applyCmd(cmd *command) (*fsmResponse, error) {
	if s.r.State() != raft.Leader {
		return nil, fmt.Errorf("unable to set on a non-leader")
	}

	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}

	f := s.r.Apply(cmdBytes, timeout)
	if err := f.Error(); err != nil {
		return nil, err
	}

	resp := f.Response().(*fsmResponse)
	return resp, nil
}

// fsm is for log replication
type fsm Raft

// Apply takes a command from the latest Log, and applies it to the store
func (sm *fsm) Apply(l *raft.Log) interface{} {
	var cmd command
	if err := json.Unmarshal(l.Data, &cmd); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command data for raft log: %s", err))
	}

	var b Bucket
	if cmd.BucketPath != "" {
		var err error
		b, err = (&Raft(sm)).bucket(cmd.BucketPath)
		if err != nil {
			return &fsmResponse{error: fmt.Errorf("unable to find bucket %s", cmd.BucketPath)}
		}
	}

	switch cmd.Op {
	case opDeleteBucket:
		err := b.DeleteBucket(string(cmd.Key))
		return &fsmResponse{error: err}
	case opDelete:
		err := b.Delete(Node{
			Key:   cmd.Key,
			Value: cmd.Value,
		})
		return &fsmResponse{error: err}
	case opSet:
		err := b.Set(Node{
			Key:   cmd.Key,
			Value: cmd.Value,
		})
		return &fsmResponse{error: err}
	case opList:
		ns, err := b.List()
		return &fsmResponse{nodes: ns, error: err}
	case opGet:
		n, err := b.Get(cmd.Key)
		return &fsmResponse{
			node: Node{
				Key:   cmd.Key,
				Value: n,
			},
			error: err,
		}
	case opSeek:
		n, err := b.Seek(cmd.Key)
		return &fsmResponse{
			node: Node{
				Key:   cmd.Key,
				Value: n,
			},
			error: err,
		}

	default:
		return &fsmResponse{error: fmt.Errorf("unknown command %v", cmd.Op)}
	}
}

// Snapshot returns an FSMSnapshot for store snapshotting
func (sm *fsm) Snapshot() (raft.FSMSnapshot, error) {
	return &fsmSnapshot{sm.s}, nil
}

// Restore restores the database from a snapshot
func (sm *fsm) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	switch store := sm.s.(type) {
	case *BoltDB:
		err := sm.s.Close()
		if err != nil {
			return err
		}

		var storeBytes []byte
		_, err = rc.Read(storeBytes)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(store.dbpath, storeBytes, 0600)
		if err != nil {
			return err
		}

		store, err = NewDB(store.dbpath)
		if err != nil {
			return nil
		}

		sm.s = Store(store)

		return nil
	case *BoltBucket:
		return fmt.Errorf("restore not possible on a *BoltBucket")
	}
	return nil
}

// fsmSnapshot represents the struct that can snapshot a database
type fsmSnapshot struct {
	s Store
}

// Persist stores the bytes of the database to the Raft sink
func (s *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	switch store := s.s.(type) {
	case *BoltDB:
		return store.db.View(func(tx *bolt.Tx) error {
			_, err := tx.WriteTo(sink)
			return err
		})
	case *BoltBucket:
		return fmt.Errorf("persist not possibly on a *BoltBucket")
	}

	return nil
}

// Release is a no-op
func (s *fsmSnapshot) Release() {}

// fsmResponse is a general response ApplyFuture type
type fsmResponse struct {
	node  Node
	nodes []Node
	error error
}
