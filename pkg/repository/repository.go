package repository

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/unerror/waffy/pkg/data"
)

func Create(b data.Store, k []byte, m proto.Marshaler) error {
	if _, err := b.Get(k); err == nil {
		return fmt.Errorf("%s already exists", k)
	}

	mBytes, err := m.Marshal()
	if err != nil {
		return err
	}

	return b.Set(data.Node{
		Key:   k,
		Value: mBytes,
	})
}
