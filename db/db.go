package db

import (
	"bytes"
	"github.com/goproxy/goproxy/constant"
	"github.com/goproxy/goproxy/export"
	"github.com/goproxy/goproxy/logger"
	"github.com/juju/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"sync"
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
		var bkt *bbolt.Bucket
		for _, bk := range [][]byte{constant.Cfg, constant.CheckPoint, constant.FS} {
			bkt, err = tx.CreateBucketIfNotExists(bk)
			if err != nil {
				return errors.Annotatef(err, "create bucket: %s", string(bk))
			}
			if bytes.Equal(bk, constant.Cfg) {
				if bkt.Get(constant.Head) == nil {
					logger.Debug("header is missing, will set be empty header id")
					err = bkt.Put(constant.Head, export.EmptyId)
					if err != nil {
						return errors.Annotate(err, "put empty root ID to header")
					}
				}
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

func View(f func(tx *bbolt.Tx) error) error {
	return db.View(f)
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
