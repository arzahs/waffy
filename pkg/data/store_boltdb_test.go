package data

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	setNodeC = Node{
		Key:   []byte("C"),
		Value: []byte("kyle"),
	}
)

func TestBoltDB(t *testing.T) {
	d, err := NewDB("test.db")

	Convey("Open on a bad path should error", t, func() {
		_, err := NewDB("..")
		So(err, ShouldBeError)
	})

	Convey("Opening a valid database should open the path", t, func() {
		So(err, ShouldBeNil)

		So(d, ShouldHaveSameTypeAs, &BoltDB{})
	})

	Convey("Creating a bucket should store a bucket in BoltDB", t, func() {
		b, err := d.Bucket("A")
		So(err, ShouldBeNil)
		So(b, ShouldHaveSameTypeAs, &BoltBucket{})
		So(b, ShouldNotBeNil)

		So(d.buckets, ShouldContainKey, "A")
		So(b, ShouldHaveSameTypeAs, &BoltBucket{})

		Convey("Create a Bucket in an Bucket should store a bucket in BoltDB", func() {
			bbStore, err := b.Bucket("B")
			bb := bbStore.(*BoltBucket)
			So(err, ShouldBeNil)
			So(bb, ShouldHaveSameTypeAs, &BoltBucket{})
			So(bb, ShouldNotBeNil)

			So(bb.parent, ShouldEqual, b)
			So(bb.parent.(*BoltBucket).buckets, ShouldContainKey, "B")
			So(bb.parent.(*BoltBucket).buckets, ShouldNotContainKey, "A")
		})

		Convey("Deleting a Bucket should remove it", func() {
			err := b.DeleteBucket("B")
			So(err, ShouldBeNil)

			So(b.(*BoltBucket).buckets, ShouldNotContainKey, "B")
			So(d.buckets, ShouldContainKey, "A")
		})

		Convey("Setting a value should be Gettable later", func() {
			err := b.Set(Node{
				Key:   []byte("waffy"),
				Value: []byte("test"),
			})
			So(err, ShouldBeNil)

			val, err := b.Get([]byte("waffy"))
			So(err, ShouldBeNil)

			So(val, ShouldResemble, []byte("test"))

			Convey("Deleteing a key should remove it", func() {
				err := b.Delete(Node{
					Key: []byte("waffy"),
				})
				So(err, ShouldBeNil)

				data, err := b.Get([]byte("waffy"))
				So(err, ShouldNotBeNil)
				So(data, ShouldBeNil)

				Convey("Deleting a key by value works", func() {
					b.Set(setNodeC)
					err := b.Delete(Node{
						Value: setNodeC.Value,
					})
					So(err, ShouldNotBeNil)
				})
			})
		})
	})

	Convey("Listening a Bucket should list Keys and Buckets", t, func() {
		b, _ := d.Bucket("root")
		b.Bucket("A")
		b.Bucket("B")
		b.Set(setNodeC)

		nodes, err := b.List()
		So(err, ShouldBeNil)
		So(nodes, ShouldNotBeEmpty)
		So(nodes, ShouldContain, Node{
			Key:    []byte("A"),
			Bucket: true,
		})
		So(nodes, ShouldContain, Node{
			Key:    []byte("B"),
			Bucket: true,
		})
		So(nodes, ShouldContain, setNodeC)
	})

	Convey("Closing the database should not error", t, func() {
		err := d.Close()
		So(err, ShouldBeNil)
	})

	defer func() {
		os.Remove("test.db")
	}()
}
