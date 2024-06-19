package export

import (
	"crypto/sha1"
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

func (x *FileEntry) ComputeId() {
	x.Id = nil
	x.Id = toId(x)
}
