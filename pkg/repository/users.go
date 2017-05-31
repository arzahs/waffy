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

// FindUserByEmail returns the User stored with the given email
func FindUserByEmail(d data.Bucket, email string) (*users.User, error) {
	b, err := d.Bucket(UsersBucket)
	if err != nil {
		return nil, err
	}

	u := users.User{}
	err = Seek(b, []byte(email), &u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}
