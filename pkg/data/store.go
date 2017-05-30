// Package data is responsible for data management within the load balancer
package data

// Node presents a single key/value pair
type Node struct {
	Key []byte
	Value []byte
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
	List() ([]Node, error)
}

type Bucket interface {
	Bucket(name string) (Store, error)
	DeleteBucket(name string) error
}

// Store represents a data store that can be loaded as key/value pairs
type Store interface {
	Bucket
	ValueLister
	ValueGetter
	ValueSetter
	ValueDeleter
}
