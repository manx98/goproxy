package export

import (
	"archive/zip"
	"context"
	"github.com/goproxy/goproxy/cache"
	"github.com/goproxy/goproxy/obj"
	"github.com/juju/errors"
	"go.etcd.io/bbolt"
	"io"
	"os"
	"path/filepath"
)

const (
	AddDirentDir    = "A"
	DeleteDirentDir = "D"
)

type DiffZipper struct {
	w   io.Writer
	zw  *zip.Writer
	st  *StreamDataWriter
	chr cache.Cacher
}

func (z *DiffZipper) writeFile(op OperateType, ctx context.Context, parent string, name string) error {
	if op == DiffStatusAdded {
		reader, err := z.chr.Get(ctx, filepath.Join(parent, name))
		if err != nil {
			return err
		}
		create, err := z.zw.Create(filepath.Join(AddDirentDir, parent, name))
		if err != nil {
			return errors.Annotate(err, "create zip file")
		}
		buf := make([]byte, 1024)
		for ctx.Err() == nil {
			readNum, err1 := reader.Read(buf)
			if readNum > 0 {
				_, err = create.Write(buf[:readNum])
				if err == nil {
					err = z.st.AddSize(int64(readNum))
				}
			}
			if err != nil {
				return errors.Annotate(err1, "write zip file")
			}
			if err1 != nil {
				if errors.Is(err1, io.EOF) {
					return nil
				}
				return errors.Annotate(err1, "read zip file")
			}
		}
		return context.Cause(ctx)
	} else {
		_, err := z.zw.Create(filepath.Join(DeleteDirentDir, parent, name))
		if err != nil {
			return errors.Annotate(err, "add delete file")
		}
	}
	return nil
}

func (z *DiffZipper) writeDir(op OperateType, ctx context.Context, parent string, name string) error {
	dirs, err := z.chr.List(ctx, filepath.Join(parent, name))
	if err != nil {
		return err
	}
	if err = z.st.AddDir(1); err != nil {
		return err
	}
	if op == DiffStatusDeleted {
		header := &zip.FileHeader{
			Name:   filepath.Join(DeleteDirentDir, parent, name),
			Method: zip.Store,
		}
		header.SetMode(os.ModeDir | 0755)
		_, err = z.zw.CreateHeader(header)
		if err != nil {
			return errors.Annotate(err, "add delete dir zip header")
		}
	} else {
		if len(dirs) <= 0 {
			header := &zip.FileHeader{
				Name:   filepath.Join(AddDirentDir, parent, name),
				Method: zip.Store,
			}
			header.SetMode(os.ModeDir | 0755)
			_, err = z.zw.CreateHeader(header)
			if err != nil {
				return errors.Annotate(err, "add add dir zip header")
			}
		}
		for _, dirent := range dirs {
			if dirent.IsDir {
				err = z.writeDir(op, ctx, filepath.Join(parent, name), dirent.Name)
			} else {
				err = z.writeFile(op, ctx, filepath.Join(parent, name), dirent.Name)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (z *DiffZipper) callback(ctx context.Context, op OperateType, parent string, dirent *obj.Dirent) error {
	if dirent.IsDir {
		return z.writeDir(op, ctx, parent, dirent.Name)
	} else {
		return z.writeFile(op, ctx, parent, dirent.Name)
	}
}

func (z *DiffZipper) Close() error {
	return z.zw.Close()
}

func (z *DiffZipper) DiffHead(ctx context.Context, tx *bbolt.Tx, oldRoot []byte) (err error) {
	err = DiffHead(oldRoot, &DiffOpt{
		Tx:  tx,
		Ctx: ctx,
		Callback: func(op OperateType, parent string, dirent *obj.Dirent) error {
			return z.callback(ctx, op, parent, dirent)
		},
		Chr: z.chr,
		Stw: z.st,
	})
	return
}

func NewDiffZipper(chr cache.Cacher, out io.Writer, st *StreamDataWriter) *DiffZipper {
	zipper := &DiffZipper{
		w:   out,
		chr: chr,
		st:  st,
	}
	zipper.zw = zip.NewWriter(st)
	return zipper
}
