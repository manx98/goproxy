package obj

import (
	"crypto/sha1"
	"github.com/juju/errors"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

func toId(m proto.Message) []byte {
	hash := sha1.New()
	marshal, err := proto.Marshal(m)
	if err != nil {
		panic(err)
	}
	hash.Write(marshal)
	return hash.Sum(nil)
}

func (x *DirEntry) ComputeId() {
	x.Id = nil
	x.Id = toId(x)
}

func (x *DirEntry) Save(bucket *bbolt.Bucket) error {
	if x.Id == nil {
		x.ComputeId()
	}
	if bucket.Get(x.Id) == nil {
		data, err := proto.Marshal(x)
		if err != nil {
			return errors.Annotate(err, "marshal dir entry")
		}
		err = bucket.Put(x.Id, data)
		return errors.Annotate(err, "put dir entry to db")
	}
	return nil
}

func (x *Dirent) ComputeId() {
	x.Id = nil
	x.Id = toId(x)
}

func (x *CheckPoint) ComputeId() {
	x.Id = nil
	x.Id = toId(x)
}

func (x *CheckPoint) Save(bucket *bbolt.Bucket) error {
	if x.Id == nil {
		x.ComputeId()
	}
	if bucket.Get(x.Id) == nil {
		data, err := proto.Marshal(x)
		if err != nil {
			return errors.Annotate(err, "marshal checkpoint")
		}
		err = bucket.Put(x.Id, data)
		return errors.Annotate(err, "put checkpoint to db")
	}
	return nil
}
