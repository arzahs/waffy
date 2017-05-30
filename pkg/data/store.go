// Package data is responsible for data management within the load balancer
package data

// Node presents a single key/value pair
type Node struct {
	Key string
	Value []byte
	Bucket bool
}

// ValueGetter is an interface that can get data
type ValueGetter interface {
	// Get returns the value for key k
	Get(k string) ([]byte, error)
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

type Buckets interface {
	Bucket(name string) Store
	DeleteBucket(name string) Store
}

// Store represents a data store that can be loaded as key/value pairs
type Store interface {
	Buckets
	ValueGetter
	ValueSetter
	ValueDeleter
}
