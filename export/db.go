package export

import (
	"github.com/goproxy/goproxy/logger"
	"github.com/juju/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"sync"
)

var (
	checkPointKey = []byte("CPK")
	cfgKey        = []byte("CFG")
	headKey       = []byte("HD")
	fsKey         = []byte("FS")
)

var (
	db   *bbolt.DB
	Lock sync.RWMutex
)

func InitDb(dataPath string) {
	var err error
	db, err = bbolt.Open(dataPath, 0600, nil)
	if err != nil {
		logger.Fatal("failed to create database", zap.Error(err))
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		for _, bk := range [][]byte{cfgKey, checkPointKey, fsKey} {
			_, err = tx.CreateBucketIfNotExists(bk)
			if err != nil {
				return errors.Annotatef(err, "create bucket: %s", string(bk))
			}
		}
		return nil
	})
	if err != nil {
		logger.Fatal("failed to create bucket", zap.Error(err))
	}
	return
}

func Update(f func(tx *bbolt.Tx) error) error {
	return db.Update(f)
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
