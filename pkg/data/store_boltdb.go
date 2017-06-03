package data

import (
	"fmt"

	"bytes"

	"github.com/boltdb/bolt"
)

// BoltDB represents a database connection to BoltDB, and underlying Store storage
type BoltDB struct {
	db      *bolt.DB
	buckets map[string]*BoltBucket
}

// BoltBucket is an implementation of Bucket using BoltDB
type BoltBucket struct {
	db      *bolt.DB
	buckets map[string]*BoltBucket
	name    []byte
	parent  Store
}

// NewDB returns a new *BoltDB Bucket
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

// Store returns a new base bucket on the BoltDB store
func (d *BoltDB) Bucket(name string) (Bucket, error) {
	bucketName := []byte(name)

	if b, ok := d.buckets[name]; ok {
		return b, nil
	}

	err := d.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)

		return err
	})
	if err != nil {
		return nil, err
	}

	bb := &BoltBucket{
		db:      d.db,
		buckets: make(map[string]*BoltBucket),
		name:    bucketName,
		parent:  d,
	}

	d.buckets[name] = bb

	return bb, err
}

// DeleteBucket removes a base bucket on the BoltDB store
func (d *BoltDB) DeleteBucket(name string) error {
	bucketName := []byte(name)

	err := d.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(bucketName)
	})
	if err != nil {
		return err
	}

	delete(d.buckets, name)

	return nil
}

// Store creates (or fetches) a BoltBucket with the given name, from this leaf BoltBucket
func (s *BoltBucket) Bucket(name string) (Bucket, error) {
	if b, ok := s.buckets[name]; ok {
		return b, nil
	}

	err := s.db.Update(func(tx *bolt.Tx) error {
		_, err := s.bucket(tx, []byte(name))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	bb := &BoltBucket{
		db:      s.db,
		buckets: make(map[string]*BoltBucket),
		name:    []byte(name),
		parent:  s,
	}

	s.buckets[name] = bb

	return bb, err
}

// Close closes the connection to the data Bucket
func (s *BoltBucket) Close() error {
	return s.db.Close()
}

func (s *BoltBucket) bucket(tx *bolt.Tx, name []byte) (*bolt.Bucket, error) {
	switch s.parent.(type) {
	case *BoltDB:
		bucket, err := tx.CreateBucketIfNotExists(s.name)
		if err != nil {
			return nil, err
		}
		return bucket.CreateBucketIfNotExists([]byte(name))
	case *BoltBucket:
		parent, err := s.parent.(*BoltBucket).bucket(tx, s.name)
		if err != nil {
			return nil, err
		}

		return parent.CreateBucketIfNotExists([]byte(name))
	}

	return nil, fmt.Errorf("unable to create bucket")
}

func (s *BoltBucket) this(tx *bolt.Tx) (*bolt.Bucket, error) {
	var parent *bolt.Bucket
	switch s.parent.(type) {
	case *BoltDB:
		parent = tx.Bucket(s.name)
	case *BoltBucket:
		var err error
		parent, err = s.parent.(*BoltBucket).bucket(tx, s.name)
		if err != nil {
			return nil, err
		}
	}

	return parent, nil
}

// DeleteBucket removes a leaf BoltBucket from this BoltBucket
func (s *BoltBucket) DeleteBucket(name string) error {
	if _, ok := s.buckets[name]; !ok {
		return fmt.Errorf("bucket does not exist")
	}

	err := s.db.Update(func(tx *bolt.Tx) error {
		var parent *bolt.Bucket
		switch s.parent.(type) {
		case *BoltDB:
			parent = tx.Bucket(s.name)
		case *BoltBucket:
			var err error
			parent, err = s.parent.(*BoltBucket).bucket(tx, s.name)
			if err != nil {
				return err
			}
		}

		return parent.DeleteBucket([]byte(name))
	})
	if err != nil {
		return err
	}

	delete(s.buckets, name)

	return nil
}

// Close closes the database connection
func (d *BoltDB) Close() error {
	return d.db.Close()
}

// Get returns the value of a key in the BoltBucket
func (s *BoltBucket) Get(k []byte) ([]byte, error) {
	var value []byte
	err := s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := s.this(tx)
		if err != nil {
			return err
		}

		value = bucket.Get(k)
		if len(value) == 0 {
			return fmt.Errorf("key %s does not exist", k)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return value, nil
}

// Set stores a Node in the BoltBucket
func (s *BoltBucket) Set(n Node) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := s.this(tx)
		if err != nil {
			return err
		}

		return bucket.Put(n.Key, n.Value)
	})
}

// Delete deletes a Node by Key or by Value. Delete by Key does a simple delete. Delete by Value
// iterates over the database for the Node with the given Value
func (s *BoltBucket) Delete(n Node) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := s.this(tx)
		if err != nil {
			return err
		}

		if n.Key != nil {
			return bucket.Delete(n.Key)
		}

		if n.Value != nil {
			bucket.ForEach(func(k, v []byte) error {
				if bytes.Equal(v, n.Value) {
					return bucket.Delete(k)
				}
				return fmt.Errorf("no value node found")
			})
		}
		return fmt.Errorf("no node deleted")
	})
}

// List returns the Nodes. Nodes with empty an Value are Store
func (s *BoltBucket) List() ([]Node, error) {
	var nodes []Node

	err := s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := s.this(tx)
		if err != nil {
			return err
		}

		return bucket.ForEach(func(k, v []byte) error {
			nodes = append(nodes, Node{
				Key:    k,
				Value:  v,
				Bucket: false,
			})

			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	for name := range s.buckets {
		nodes = append(nodes, Node{
			Key:    []byte(name),
			Bucket: true,
		})
	}

	return nodes, nil
}

// Seek seeks a given key k in the Store
func (s *BoltBucket) Seek(k []byte) ([]byte, error) {
	var value []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket, err := s.this(tx)
		if err != nil {
			return err
		}

		c := bucket.Cursor()
		var key []byte
		key, value = c.Seek(k)
		if key == nil {
			return fmt.Errorf("key %s not found", k)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return value, err
}
