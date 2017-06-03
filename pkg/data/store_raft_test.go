package data

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRaftStore(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("", "raft_test")
	defer os.Remove(tmpDir)

	s, _ := NewDB(fmt.Sprintf("%s/raft_test.db", tmpDir))
	d, err := NewRaft(tmpDir, "127.0.0.1:0", s)

	Convey("Opening a bad Raft listen address should fail", t, func() {
		_, err := NewRaft(tmpDir, "256.0.0.0:3453", s)
		So(err, ShouldNotBeNil)
	})

	Convey("Opening a database should open the consensus", t, func() {
		So(err, ShouldBeNil)
		So(d, ShouldHaveSameTypeAs, &Raft{})
	})

	Convey("Creating a bucket should create a bucket on the single Leader", t, func() {
		b, err := d.Bucket("A")
		So(err, ShouldBeNil)
		So(b, ShouldHaveSameTypeAs, &Raft{})

		So(b.(*Raft).path, ShouldEqual, "/A/")
		underlyingBucketA, err := b.(*Raft).bucket("/A/")
		So(err, ShouldBeNil)
		So(underlyingBucketA, ShouldHaveSameTypeAs, &BoltBucket{})

		Convey("Creating a Bucket in a Store", func() {
			bbStore, err := b.Bucket("B")
			So(err, ShouldBeNil)
			So(bbStore, ShouldHaveSameTypeAs, &Raft{})
			So(bbStore.(*Raft).path, ShouldEqual, "/A/B/")

			So(underlyingBucketA.(*BoltBucket).buckets, ShouldContainKey, "B")

			underlyingBucketB, err := b.(*Raft).bucket("/A/B/")
			So(underlyingBucketB, ShouldHaveSameTypeAs, &BoltBucket{})
			So(underlyingBucketB.(*BoltBucket).parent, ShouldEqual, underlyingBucketA)
		})

		Convey("Deleting a Bucket should remove it from the Store", func() {
			err := b.DeleteBucket("B")
			So(err, ShouldBeNil)

			So(underlyingBucketA.(*BoltBucket).buckets, ShouldNotContainKey, "B")
		})

		Convey("Setting a value should be Gettable leader", func() {
			err := b.Set(Node{
				Key:   []byte("waffy"),
				Value: []byte("test"),
			})
			So(err, ShouldBeNil)

			Convey("GetWeak should return a weakly consistent value", func() {
				val, err := b.(Consensus).GetWeak([]byte("waffy"))
				So(err, ShouldBeNil)
				So(val, ShouldResemble, []byte("test"))
			})

			Convey("Get should return a strongly consistent value", func() {
				val, err := b.Get([]byte("waffy"))
				So(err, ShouldBeNil)
				So(val, ShouldResemble, []byte("test"))
			})

			Convey("Deleting a key should remove it form the underlying bucket", func() {
				err := b.Delete(Node{
					Key: []byte("waffy"),
				})
				So(err, ShouldBeNil)

				data, err := b.Get([]byte("waffy"))
				So(err, ShouldNotBeNil)
				So(data, ShouldBeNil)

				underData, err := underlyingBucketA.Get([]byte("waffy"))
				So(err, ShouldNotBeNil)
				So(underData, ShouldBeNil)
			})

			Convey("Deleting a key by value works", func() {
				b.Set(Node{
					Key:   []byte("waffy"),
					Value: []byte("waffy"),
				})
				err := b.Delete(Node{
					Value: []byte("waffy"),
				})
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Listening a Bucket should list Keys and Buckets", t, func() {
		b, _ := d.Bucket("root")
		b.Bucket("A")
		b.Bucket("B")
		b.Set(Node{
			Key:   []byte("C"),
			Value: []byte("kyle"),
		})

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
		So(nodes, ShouldContain, Node{
			Key:   []byte("C"),
			Value: []byte("kyle"),
		})

		underlyingBucket, _ := d.(*Raft).bucket("/root/")
		underNodes, _ := underlyingBucket.List()

		So(nodes, ShouldResemble, underNodes)
	})
}
