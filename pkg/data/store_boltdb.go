package data

import (
	"bytes"
	"fmt"

	"github.com/boltdb/bolt"
)

type bucketable interface {
	CreateBucket(key []byte) (*bolt.Bucket, error)
	CreateBucketIfNotExists(key []byte) (*bolt.Bucket, error)
	DeleteBucket(key []byte) error
}

// BoltDB represents a database connection to BoltDB, and underlying Bucket storage
type BoltDB struct {
	db      *bolt.DB
	buckets map[string]*BoltBucket
}

// BoltBucket is an implementation of Store using BoltDB
type BoltBucket struct {
	db      *bolt.DB
	b       *bolt.Bucket
	buckets map[string]*BoltBucket
}

// NewDB returns a new *BoltDB Store
func NewDB(path string) (*BoltDB, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	b := &BoltDB{
		db:      db,
		buckets: make(map[string]*BoltBucket),
	}

	return b, nil
}

// Bucket returns a new base bucket on the BoltDB store
func (d *BoltDB) Bucket(name string) (*BoltBucket, error) {
	if b, ok := d.buckets[name]; ok {
		return b, nil
	}

	bb, err := getOrCreateBucket(d.db, &bolt.Tx{}, []byte(name))
	if err != nil {
		return nil, fmt.Errorf("cannot create bucket: %s", err)
	}

	d.buckets[name] = bb

	return bb, err
}

// DeleteBucket removes a base bucket on the BoltDB store
func (d *BoltDB) DeleteBucket(name string) error {
	return deleteBucket(d.buckets, d.db, &bolt.Tx{}, []byte(name))
}

// Close closes the database connection
func (d *BoltDB) Close() error {
	return d.db.Close()
}

// Get returns the value of a key in the BoltBucket
func (s *BoltBucket) Get(k string) ([]byte, error) {
	var v []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		v = s.b.Get([]byte(k))

		return nil
	})
	if err != nil {
		return nil, err
	}

	return v, err
}

// Set stores a Node in the BoltBucket
func (s *BoltBucket) Set(n Node) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return s.b.Put([]byte(n.Key), []byte(n.Value))
	})
}

// Delete deletes a Node by Key or by Value. Delete by Key does a simple delete. Delete by Value
// iterates over the database for the Node with the given Value
func (s *BoltBucket) Delete(n Node) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		if err := s.b.Delete([]byte(n.Key)); err != nil {
			return err
		}

		err := s.b.ForEach(func(k, v []byte) error {
			if bytes.Equal(v, n.Value) {
				return s.b.Delete(k)
			}

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
}

// List returns the Nodes. Nodes with empty an Value are Buckets
func (s *BoltBucket) List() ([]Node, error) {
	var nodes []Node
	err := s.db.View(func(tx *bolt.Tx) error {
		return s.b.ForEach(func(k, v []byte) error {
			nodes = append(nodes, Node{
				Key:    string(k),
				Value:  v,
				Bucket: false,
			})

			return nil
		})
	})

	for name := range s.buckets {
		nodes = append(nodes, Node{
			Key:    name,
			Bucket: true,
		})
	}

	return nodes, err
}

// Bucket creates (or fetches) a BoltBucket with the given name, from this leaf BoltBucket
func (s *BoltBucket) Bucket(name string) (*BoltBucket, error) {
	if b, ok := s.buckets[name]; ok {
		return b, nil
	}

	bb, err := getOrCreateBucket(s.db, s.b, []byte(name))
	if err != nil {
		return nil, err
	}

	s.buckets[name] = bb

	return bb, nil
}

// DeleteBucket removes a leaf BoltBucket from this BoltBucket
func (s *BoltBucket) DeleteBucket(name string) error {
	return deleteBucket(s.buckets, s.db, s.b, []byte(name))
}

func getOrCreateBucket(db *bolt.DB, bable bucketable, k []byte) (*BoltBucket, error) {
	var b *bolt.Bucket

	err := db.Update(func(tx *bolt.Tx) error {
		var err error
		switch buck := bable.(type) {
		case *bolt.Tx:
			b, err = tx.CreateBucketIfNotExists(k)
		default:
			b, err = buck.CreateBucketIfNotExists(k)
		}
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	bb := &BoltBucket{
		db:      db,
		b:       b,
		buckets: make(map[string]*BoltBucket),
	}

	return bb, nil
}

func deleteBucket(buckets map[string]*BoltBucket, db *bolt.DB, bable bucketable, k []byte) error {
	if _, ok := buckets[string(k)]; !ok {
		return fmt.Errorf("key %s does not exist", k)
	}

	return db.Update(func(tx *bolt.Tx) error {
		switch buck := bable.(type) {
		case *bolt.Tx:
			return tx.DeleteBucket(k)
		default:
			return buck.DeleteBucket(k)
		}
	})

	delete(buckets, string(k))

	return nil
}
