package repository

import (
	"github.com/unerror/waffy/pkg/data"
	"github.com/unerror/waffy/pkg/services/protos/users"
)

const (
	// UsersBucket is the Store Bucket that users are stored in
	UsersBucket = "users"
)

// CreateUser creates a user u in the data store d
func CreateUser(d data.Bucket, u *users.User) error {
	b, err := d.Bucket(UsersBucket)
	if err != nil {
		return err
	}

	return Create(b, []byte(u.Email), u)
}
