// Package data is responsible for data management within the load balancer
package data

// Node presents a single key/value pair
type Node struct {
	Key    []byte
	Value  []byte
	Bucket bool
}

// ValueGetter is an interface that can get data
type ValueGetter interface {
	// Get returns the value for key k
	Get(k []byte) ([]byte, error)
}

// ValueSetter is an interface that can set data
type ValueSetter interface {
	// Set sets the key k to the value v
	Set(n Node) error
}

// ValueDeleter is an interface that can delete data
type ValueDeleter interface {
	// Delete deletes
	Delete(n Node) error
}

// ValueLister is an interface that can list data
type ValueLister interface {
	// List lists the Nodes in the Bucket
	List() ([]Node, error)
}

// ValueFinder is an interface that finds data in the store
type ValueFinder interface {
	Seek(k []byte) ([]byte, error)
}

type Store interface {
	Bucket(name string) (Bucket, error)
	DeleteBucket(name string) error

	// Close closes the connection to the Data store
	Close() error
}

// Bucket represents a data store that can be loaded as key/value pairs
type Bucket interface {
	Store
	ValueLister
	ValueGetter
	ValueSetter
	ValueDeleter
	ValueFinder
}
