package export

import (
	"github.com/goproxy/goproxy/logger"
	"github.com/juju/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"path/filepath"
)

var (
	bucketKey = []byte("BUCKET")
	headKey   = []byte("HEAD")
)

var db *bbolt.DB

func initDb(dataPath string) {
	var err error
	db, err = bbolt.Open(filepath.Join(dataPath, "data.db"), 0600, nil)
	if err != nil {
		logger.Fatal("failed to create database", zap.Error(err))
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists(bucketKey)
		return err
	})
	if err != nil {
		logger.Fatal("failed to create bucket", zap.Error(err))
	}
	return
}

func UpdateHead(tx *bbolt.Tx, id []byte) error {
	return errors.Annotate(tx.Bucket(bucketKey).Put(headKey, id), "update head")
}
