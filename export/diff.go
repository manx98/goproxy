package export

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"github.com/goproxy/goproxy/cache"
	"github.com/goproxy/goproxy/constant"
	"github.com/goproxy/goproxy/db"
	"github.com/goproxy/goproxy/obj"
	"github.com/goproxy/goproxy/utils"
	"github.com/juju/errors"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"
)

type OperateType rune

const (
	DiffStatusAdded    OperateType = 'A'
	DiffStatusDeleted  OperateType = 'D'
	DiffStatusModified OperateType = 'M'
)

const (
	DirAddUpdated  byte = 'D'
	SizeAddUpdated byte = 'S'
	StatusUpdated  byte = 'M'
	BinaryWrite    byte = 'W'
)

func direntCompareFunc(a, b *obj.Dirent) int {
	return strings.Compare(a.Name, b.Name)
}

type StreamDataWriter struct {
	w         io.Writer
	available bool
	flush     func()
}

func (s *StreamDataWriter) AddDir(num int64) (err error) {
	if s.available {
		defer s.flush()
		data := make([]byte, 9)
		data[0] = DirAddUpdated
		binary.BigEndian.PutUint64(data[1:], uint64(num))
		_, err = s.w.Write(data)
		if err != nil {
			return errors.Annotate(err, "write dir num")
		}
	}
	return nil
}

func (s *StreamDataWriter) AddSize(size int64) (err error) {
	if s.available {
		defer s.flush()
		data := make([]byte, 9)
		data[0] = SizeAddUpdated
		binary.BigEndian.PutUint64(data[1:], uint64(size))
		_, err = s.w.Write(data)
		if err != nil {
			return errors.Annotate(err, "write file size")
		}
	}
	return nil
}

func (s *StreamDataWriter) Write(data []byte) (n int, err error) {
	if s.available {
		defer s.flush()
		dataPackage := make([]byte, len(data)+9)
		dataPackage[0] = BinaryWrite
		binary.BigEndian.PutUint64(dataPackage[1:], uint64(len(data)))
		copy(dataPackage[9:], data)
		n, err = s.w.Write(dataPackage)
		if err != nil {
			return 0, errors.Annotate(err, "write binary")
		}
	}
	return len(data), nil
}

func (s *StreamDataWriter) Close(msg string) (err error) {
	if s.available {
		defer s.flush()
		data := make([]byte, len(msg)+9)
		data[0] = StatusUpdated
		binary.BigEndian.PutUint64(data[1:], uint64(len(msg)))
		copy(data[9:], msg)
		_, err = s.w.Write(data)
		if err != nil {
			return errors.Annotate(err, "write msg")
		}
	}
	return nil
}

func NewCreateCheckPointWatcher(writer http.ResponseWriter) *StreamDataWriter {
	w := &StreamDataWriter{
		w:         writer,
		available: !utils.IsNil(writer),
	}
	flusher, ok := writer.(http.Flusher)
	if ok {
		w.flush = flusher.Flush
	} else {
		w.flush = func() {

		}
	}
	return w
}

type DiffCallback func(op OperateType, parent string, dirent *obj.Dirent) error

func buildNewTree(bucket *bbolt.Bucket, ctx context.Context, cher cache.Cacher, name string, st *StreamDataWriter) (id []byte, err error) {
	dirent := new(obj.DirEntry)
	dirent.Entries, err = cher.List(ctx, name)
	if err != nil {
		return nil, err
	}
	if err = st.AddDir(1); err != nil {
		return nil, err
	}
	if len(dirent.Entries) == 0 {
		return constant.EmptyId, nil
	}
	slices.SortFunc(dirent.Entries, direntCompareFunc)
	for _, info := range dirent.Entries {
		if ctx.Err() != nil {
			return nil, errors.Annotate(context.Cause(ctx), "walk dirent entry")
		}
		if info.IsDir {
			info.Mtime = 0
			info.Id, err = buildNewTree(bucket, ctx, cher, name+info.Name+"/", st)
			if err != nil {
				return
			}
		} else {
			if err = st.AddSize(info.Size); err != nil {
				return nil, err
			}
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
	if id == nil {
		return nil, errors.Errorf("CheckPoint id is nil")
	}
	if bytes.Equal(id, constant.EmptyId) {
		return nil, nil
	}
	bucket := tx.Bucket(constant.CheckPoint)
	if bucket == nil {
		return nil, errors.Errorf("bucket %s is missing", string(constant.CheckPoint))
	}
	data := bucket.Get(id)
	if data == nil {
		return nil, errors.Errorf("checkponit %s data is missing", hex.EncodeToString(data))
	}
	info := &obj.CheckPoint{Parent: constant.EmptyId, Id: constant.EmptyId}
	err := proto.Unmarshal(data, info)
	if err != nil {
		return nil, errors.Annotatef(err, "unmarshal checkpoint %s data", hex.EncodeToString(data))
	}
	return info, nil
}

func GetHead(tx *bbolt.Tx) (*obj.CheckPoint, error) {
	bucket := tx.Bucket(constant.Cfg)
	if bucket == nil {
		return nil, errors.Errorf("bucket %s is missing", string(constant.Cfg))
	}
	data := bucket.Get(constant.Head)
	if data == nil {
		return nil, errors.Errorf("head is missing")
	}
	return GetCheckPoint(tx, data)
}

func SetHead(tx *bbolt.Tx, id []byte) error {
	bucket := tx.Bucket(constant.Cfg)
	if bucket == nil {
		return errors.Errorf("bucket %s is missing", string(constant.Cfg))
	}
	if err := bucket.Put(constant.Head, id); err != nil {
		return errors.Annotate(err, "put head")
	}
	return nil
}

func BuildNewTree(tx *bbolt.Tx, ctx context.Context, cher cache.Cacher, st *StreamDataWriter) ([]byte, error) {
	bucket := tx.Bucket(constant.FS)
	if bucket == nil {
		return nil, errors.Errorf("bucket %s is missing", string(constant.FS))
	}
	return buildNewTree(bucket, ctx, cher, "", st)
}

func GetDirEntry(tx *bbolt.Tx, id []byte) (*obj.DirEntry, error) {
	info := &obj.DirEntry{Id: constant.EmptyId}
	if bytes.Equal(id, constant.EmptyId) {
		return info, nil
	}
	bucket := tx.Bucket(constant.FS)
	if bucket == nil {
		return nil, errors.Errorf("bucket %s is missing", string(constant.FS))
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

func CreateCheckPoint(tx *bbolt.Tx, ctx context.Context, desc string, cher cache.Cacher, st *StreamDataWriter) ([]byte, error) {
	db.Lock.Lock()
	defer db.Lock.Unlock()
	head, err := GetHead(tx)
	if err != nil {
		return nil, err
	}
	newHeader := &obj.CheckPoint{
		Parent: constant.EmptyId,
		Desc:   desc,
		Mtime:  time.Now().UnixMilli()}
	if head != nil {
		newHeader.Parent = head.Id
	}
	newHeader.Id, err = BuildNewTree(tx, ctx, cher, st)
	if err != nil {
		return nil, err
	}
	if bytes.Compare(newHeader.Parent, newHeader.Id) == 0 {
		return nil, nil
	}
	bucket := tx.Bucket(constant.CheckPoint)
	if bucket == nil {
		return nil, errors.Errorf("bucket %s is missing", string(constant.CheckPoint))
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
	Chr      cache.Cacher
	Tx       *bbolt.Tx
	Ctx      context.Context
	Callback DiffCallback
	Stw      *StreamDataWriter
}

func (d *DiffOpt) ReadDir(name string) (*obj.DirEntry, error) {
	dirs, err := d.Chr.List(d.Ctx, name)
	if err != nil {
		return nil, err
	}
	if err = d.Stw.AddDir(1); err != nil {
		return nil, err
	}
	for _, dir := range dirs {
		if !dir.IsDir {
			dir.ComputeId()
		} else if err = d.Stw.AddSize(dir.Size); err != nil {
			return nil, err
		}
	}
	slices.SortFunc(dirs, direntCompareFunc)
	return &obj.DirEntry{
		Entries: dirs,
	}, nil
}

func direntSame(dentA, dentB *obj.Dirent) bool {
	return !dentA.IsDir && bytes.Equal(dentA.Id, dentB.Id)
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

	return twoWayDiffFiles(baseDir, files, opt)
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
	err := twoWayDiffDirs(baseDir, dirs, opt, &recurse)
	if err != nil {
		return err
	}
	if !recurse {
		return nil
	}
	var dirName string
	for i := 0; i < n; i++ {
		if dents[i] != nil && dents[i].IsDir {
			if i == 0 {
				subDirs[i], err = GetDirEntry(opt.Tx, dents[i].Id)
			} else {
				subDirs[i], err = opt.ReadDir(baseDir + dents[i].Name)
			}
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
		*recurse = false
		return opt.Callback(DiffStatusAdded, baseDir, p2)
	}

	if p2 == nil {
		*recurse = false
		return opt.Callback(DiffStatusDeleted, baseDir, p1)
	}

	return nil
}

func DiffTrees(oldRoot []byte, currentPath string, opt *DiffOpt) (err error) {
	trees := make([]*obj.DirEntry, 2)
	trees[0], err = GetDirEntry(opt.Tx, oldRoot)
	if err == nil {
		trees[1], err = opt.ReadDir(currentPath)
	}
	if err == nil {
		err = diffTreesRecursive(trees, currentPath, opt)
	}
	return
}

func DiffHead(oldRoot []byte, opt *DiffOpt) error {
	return DiffTrees(oldRoot, "", opt)
}
