package export

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"github.com/goproxy/goproxy/cache"
	"github.com/goproxy/goproxy/obj"
	"github.com/juju/errors"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"time"
)

var EmptyId = make([]byte, sha1.Size)

type OperateType rune

const (
	DiffStatusAdded      = 'A'
	DiffStatusDeleted    = 'D'
	DiffStatusModified   = 'M'
	DiffStatusDirAdded   = 'B'
	DiffStatusDirDeleted = 'C'
)

type CreateCheckPointStatistic struct {
	Dirs  atomic.Int64
	Files atomic.Int64
	Size  atomic.Int64
	Msg   atomic.Value
}

type DiffCallback func(op OperateType, parent string, dirent *obj.Dirent) error

func buildNewTree(bucket *bbolt.Bucket, ctx context.Context, cher cache.Cacher, name string, st *CreateCheckPointStatistic) (id []byte, err error) {
	dirent := new(obj.DirEntry)
	dirent.Entries, err = cher.List(ctx, name)
	if err != nil {
		return nil, err
	}
	st.Dirs.Add(1)
	if len(dirent.Entries) == 0 {
		return EmptyId, nil
	}
	slices.SortFunc(dirent.Entries, func(a, b *obj.Dirent) int {
		return strings.Compare(a.Name, b.Name)
	})
	for _, info := range dirent.Entries {
		if ctx.Err() != nil {
			return nil, errors.Annotate(context.Cause(ctx), "walk dirent entry")
		}
		if info.IsDir {
			info.Mtime = 0
			info.Id, err = buildNewTree(bucket, ctx, cher, filepath.Join(name, info.Name), st)
			if err != nil {
				return
			}
		} else {
			st.Files.Add(1)
			st.Size.Add(info.Size)
			info.ComputeId()
		}
	}
	err = dirent.Save(bucket)
	if err != nil {
		return
	}
	return dirent.Id, nil
}

func GetCheckPoint(tx *bbolt.Tx, id []byte) (*obj.CheckPoint, error) {
	info := &obj.CheckPoint{Parent: EmptyId, Id: EmptyId}
	if id == nil || bytes.Equal(id, EmptyId) {
		return info, nil
	}
	bucket := tx.Bucket(checkPointKey)
	if bucket == nil {
		return nil, errors.Errorf("bucket %s is missing", string(checkPointKey))
	}
	data := bucket.Get(id)
	if data == nil {
		return nil, errors.Errorf("checkponit %s data is missing", hex.EncodeToString(data))
	}
	err := proto.Unmarshal(data, info)
	if err != nil {
		return nil, errors.Annotatef(err, "unmarshal checkpoint %s data", hex.EncodeToString(data))
	}
	return info, nil
}

func GetHead(tx *bbolt.Tx) (*obj.CheckPoint, error) {
	bucket := tx.Bucket(cfgKey)
	if bucket == nil {
		return nil, errors.Errorf("bucket %s is missing", string(cfgKey))
	}
	data := bucket.Get(headKey)
	return GetCheckPoint(tx, data)
}

func SetHead(tx *bbolt.Tx, id []byte) error {
	bucket := tx.Bucket(cfgKey)
	if bucket == nil {
		return errors.Errorf("bucket %s is missing", string(cfgKey))
	}
	if err := bucket.Put(headKey, id); err != nil {
		return errors.Annotate(err, "put head")
	}
	return nil
}

func BuildNewTree(tx *bbolt.Tx, ctx context.Context, cher cache.Cacher, st *CreateCheckPointStatistic) ([]byte, error) {
	bucket := tx.Bucket(fsKey)
	if bucket == nil {
		return nil, errors.Errorf("bucket %s is missing", string(fsKey))
	}
	return buildNewTree(bucket, ctx, cher, "", st)
}

func GetDirEntry(tx *bbolt.Tx, id []byte) (*obj.DirEntry, error) {
	info := &obj.DirEntry{Id: EmptyId}
	if bytes.Equal(id, EmptyId) {
		return info, nil
	}
	bucket := tx.Bucket(fsKey)
	if bucket == nil {
		return nil, errors.Errorf("bucket %s is missing", string(fsKey))
	}
	data := bucket.Get(id)
	if data == nil {
		return nil, errors.Errorf("dir entry %s is missing", hex.EncodeToString(id))
	}
	if err := proto.Unmarshal(data, info); err != nil {
		return nil, errors.Annotatef(err, "unmarsh dir entry %s", hex.EncodeToString(id))
	}
	return info, nil
}

func CreateCheckPoint(tx *bbolt.Tx, ctx context.Context, desc string, cher cache.Cacher, st *CreateCheckPointStatistic) ([]byte, error) {
	Lock.Lock()
	defer Lock.Unlock()
	head, err := GetHead(tx)
	if err != nil {
		return nil, err
	}
	newHeader := &obj.CheckPoint{
		Parent: head.Id,
		Desc:   desc,
		Mtime:  time.Now().UnixMilli()}
	newHeader.Id, err = BuildNewTree(tx, ctx, cher, st)
	if err != nil {
		return nil, err
	}
	if bytes.Compare(newHeader.Parent, newHeader.Id) == 0 {
		return nil, nil
	}
	bucket := tx.Bucket(checkPointKey)
	if bucket == nil {
		return nil, errors.Errorf("bucket %s is missing", string(checkPointKey))
	}
	err = newHeader.Save(bucket)
	if err != nil {
		return nil, err
	}
	err = SetHead(tx, newHeader.Id)
	if err != nil {
		return nil, err
	}
	return newHeader.Id, nil
}

type DiffOpt struct {
	Tx          *bbolt.Tx
	Ctx         context.Context
	Callback    DiffCallback
	FoldDirDiff bool
	fileCB      func(baseDir string, files []*obj.Dirent, opt *DiffOpt) error
	dirCB       func(baseDir string, dirs []*obj.Dirent, opt *DiffOpt, recurse *bool) error
}

func direntSame(dentA, dentB *obj.Dirent) bool {
	return bytes.Equal(dentA.Id, dentB.Id)
}

func diffFiles(baseDir string, dents []*obj.Dirent, opt *DiffOpt) error {
	n := len(dents)
	var nFiles int
	files := make([]*obj.Dirent, 3)
	for i := 0; i < n; i++ {
		if dents[i] != nil && !dents[i].IsDir {
			files[i] = dents[i]
			nFiles++
		}
	}

	if nFiles == 0 {
		return nil
	}

	return opt.fileCB(baseDir, files, opt)
}

func diffDirectories(baseDir string, dents []*obj.Dirent, opt *DiffOpt) error {
	n := len(dents)
	dirs := make([]*obj.Dirent, 3)
	subDirs := make([]*obj.DirEntry, 3)
	var nDirs int
	for i := 0; i < n; i++ {
		if dents[i] != nil && dents[i].IsDir {
			dirs[i] = dents[i]
			nDirs++
		}
	}
	if nDirs == 0 {
		return nil
	}

	recurse := true
	err := opt.dirCB(baseDir, dirs, opt, &recurse)
	if err != nil {
		return err
	}

	if !recurse {
		return nil
	}

	var dirName string
	for i := 0; i < n; i++ {
		if dents[i] != nil && dents[i].IsDir {
			subDirs[i], err = GetDirEntry(opt.Tx, dents[i].Id)
			if err != nil {
				return err
			}
			dirName = dents[i].Name
		}
	}

	newBaseDir := baseDir + dirName + "/"
	return diffTreesRecursive(subDirs, newBaseDir, opt)
}

func diffTreesRecursive(trees []*obj.DirEntry, baseDir string, opt *DiffOpt) error {
	n := len(trees)
	ptrs := make([][]*obj.Dirent, 3)

	for i := 0; i < n; i++ {
		if trees[i] != nil {
			ptrs[i] = trees[i].Entries
		} else {
			ptrs[i] = nil
		}
	}

	var firstName string
	var done bool
	var offset = make([]int, n)
	for {
		dents := make([]*obj.Dirent, 3)
		firstName = ""
		done = true
		for i := 0; i < n; i++ {
			if len(ptrs[i]) > offset[i] {
				done = false
				dent := ptrs[i][offset[i]]

				if firstName == "" {
					firstName = dent.Name
				} else if strings.Compare(dent.Name, firstName) < 0 {
					firstName = dent.Name
				}
			}
		}
		if done {
			break
		}
		for i := 0; i < n; i++ {
			if len(ptrs[i]) > offset[i] {
				dent := ptrs[i][offset[i]]
				if firstName == dent.Name {
					dents[i] = dent
					offset[i]++
				}
			}
		}

		if n == 2 && dents[0] != nil && dents[1] != nil &&
			direntSame(dents[0], dents[1]) {
			continue
		}
		if n == 3 && dents[0] != nil && dents[1] != nil &&
			dents[2] != nil && direntSame(dents[0], dents[1]) &&
			direntSame(dents[0], dents[2]) {
			continue
		}

		if err := diffFiles(baseDir, dents, opt); err != nil {
			return err
		}
		if err := diffDirectories(baseDir, dents, opt); err != nil {
			return err
		}
	}
	return nil
}

func twoWayDiffFiles(baseDir string, dents []*obj.Dirent, opt *DiffOpt) error {
	p1 := dents[0]
	p2 := dents[1]
	if p1 == nil {
		return opt.Callback(DiffStatusAdded, baseDir, p2)
	}
	if p2 == nil {
		return opt.Callback(DiffStatusDeleted, baseDir, p1)
	}
	if !direntSame(p1, p2) {
		return opt.Callback(DiffStatusModified, baseDir, p2)
	}
	return nil
}

func twoWayDiffDirs(baseDir string, dents []*obj.Dirent, opt *DiffOpt, recurse *bool) error {
	p1 := dents[0]
	p2 := dents[1]
	if p1 == nil {
		if bytes.Equal(EmptyId, p2.Id) || opt.FoldDirDiff {
			*recurse = false
			return opt.Callback(DiffStatusDirAdded, baseDir, p2)
		} else {
			*recurse = true
		}
		return nil
	}

	if p2 == nil {
		if opt.FoldDirDiff {
			*recurse = false
		} else {
			*recurse = true
		}
		return opt.Callback(DiffStatusDirDeleted, baseDir, p1)
	}

	return nil
}

func DiffTrees(tx *bbolt.Tx, oldRoot, newRoot []byte, opt *DiffOpt) (err error) {
	trees := make([]*obj.DirEntry, 2)
	trees[0], err = GetDirEntry(tx, oldRoot)
	if err == nil {
		trees[1], err = GetDirEntry(tx, newRoot)
	}
	if err == nil {
		opt.fileCB = twoWayDiffFiles
		opt.dirCB = twoWayDiffDirs
		err = diffTreesRecursive(trees, "", opt)
	}
	return
}
