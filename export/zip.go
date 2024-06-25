package export

import (
	"archive/zip"
	"context"
	"github.com/goproxy/goproxy/cache"
	"github.com/goproxy/goproxy/obj"
	"github.com/juju/errors"
	"go.etcd.io/bbolt"
	"io"
)

type DiffZipper struct {
	w   io.Writer
	zw  *zip.Writer
	st  *StreamDataWriter
	chr cache.Cacher
}

func (z *DiffZipper) writeFile(ctx context.Context, filePath string) error {
	reader, err := z.chr.Get(ctx, filePath)
	if err != nil {
		return err
	}
	create, err := z.zw.Create(filePath)
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
}

func (z *DiffZipper) writeDir(ctx context.Context, parent string, name string) error {
	dirPath := parent + name
	dirs, err := z.chr.List(ctx, dirPath)
	if err != nil {
		return err
	}
	if err = z.st.AddDir(1); err != nil {
		return err
	}
	dirPath += "/"
	for _, dirent := range dirs {
		if dirent.IsDir {
			err = z.writeDir(ctx, dirPath, dirent.Name)
		} else {
			err = z.writeFile(ctx, dirPath+dirent.Name)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (z *DiffZipper) callback(ctx context.Context, parent string, dirent *obj.Dirent) error {
	if dirent.IsDir {
		return z.writeDir(ctx, parent, dirent.Name)
	} else {
		return z.writeFile(ctx, parent+dirent.Name)
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
			if op == DiffStatusDeleted {
				return context.Cause(ctx)
			}
			return z.callback(ctx, parent, dirent)
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
