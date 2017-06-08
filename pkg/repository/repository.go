package repository

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/unerror/waffy/pkg/data"
)

// Create will create a generic Marshable message m with key k in data.Bucket b
func Create(b data.Bucket, k []byte, m proto.Marshaler) error {
	if _, err := b.Get(k); err == nil {
		return fmt.Errorf("%s already exists", k)
	}

	return Set(b, k, m)
}

// Set will set a key to a generic Marshable message m
func Set(b data.ValueSetter, k []byte, m proto.Marshaler) error {
	mBytes, err := m.Marshal()
	if err != nil {
		return err
	}

	return b.Set(data.Node{
		Key:   k,
		Value: mBytes,
	})
}

// Seek finds the Unmarshable message with the key k in the data.Bucket b
func Seek(b data.ValueFinder, k []byte, u proto.Unmarshaler) error {
	mBytes, err := b.Seek(k)
	if err != nil {
		return err
	}

	return u.Unmarshal(mBytes)
}

// DeleteByKey removes a Node in the bucket b my key k
func DeleteByKey(b data.ValueDeleter, k []byte) error {
	return b.Delete(data.Node{
		Key: k,
	})
}
