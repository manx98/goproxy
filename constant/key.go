package constant

import "crypto/sha1"

var (
	CheckPoint = []byte("CPK")
	Cfg        = []byte("CFG")
	FS         = []byte("FS")
	Head       = []byte("HEAD")
	EmptyId    = make([]byte, sha1.Size)
)

const (
	CacheDir   = "tmp"
	DbFileName = "db.bbolt"
)
