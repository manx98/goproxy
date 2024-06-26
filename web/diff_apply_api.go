package web

import (
	"archive/zip"
	"encoding/json"
	"github.com/goproxy/goproxy"
	"github.com/goproxy/goproxy/constant"
	"github.com/goproxy/goproxy/db"
	"github.com/goproxy/goproxy/export"
	"github.com/goproxy/goproxy/logger"
	"github.com/goproxy/goproxy/utils"
	"github.com/juju/errors"
	"github.com/mazrean/formstream"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
)

type WriteFunc func(p []byte) (n int, err error)

func (r WriteFunc) Write(p []byte) (n int, err error) {
	return r(p)
}

type DiffApplyStatusSender struct {
	*export.StreamDataWriter
}

func (d *DiffApplyStatusSender) SendStatus(st *DiffApplyInfo) error {
	marshal, err := json.Marshal(st)
	if err != nil {
		return errors.Annotate(err, "marshal status info")
	}
	_, err = d.Write(marshal)
	if err != nil {
		return errors.Annotate(err, "write status info")
	}
	return nil
}

func (d *DiffApplyStatusSender) write(p []byte) (n int, err error) {
	n = len(p)
	return 0, d.AddSize(int64(n))
}

func (d *DiffApplyStatusSender) WatchReader(r io.Reader) io.Reader {
	return io.TeeReader(r, WriteFunc(d.write))
}

type DiffApplyInfo struct {
	TotalSize int64 `json:"total_size"`
	Num       int   `json:"num"`
}

func applyDiffFile(w http.ResponseWriter, r *http.Request, g *goproxy.Goproxy) (err error) {
	db.Lock.Lock()
	defer db.Lock.Unlock()
	watcher := &DiffApplyStatusSender{StreamDataWriter: export.NewCreateCheckPointWatcher(w)}
	defer func() {
		if err != nil {
			if err1 := watcher.Close(err.Error()); err1 != nil {
				logger.Warn("failed to finish apply diff with error", zap.NamedError("occur_error", err), zap.Error(err1))
			}
		} else {
			if err = watcher.Close(""); err != nil {
				logger.Warn("failed to finish apply diff", zap.Error(err))
			}
		}
	}()
	var boundary string
	boundary, err = utils.GetBoundary(r.Header.Get("Content-Type"))
	if err != nil {
		return errors.Annotate(err, "get boundary")
	}
	parser := formstream.NewParser(boundary)
	var tempDir string
	tempDir, err = os.MkdirTemp(g.TempDir, constant.TempDirPattern)
	if err != nil {
		return errors.Annotate(err, "create temp dir")
	}
	defer func() {
		if err1 := os.RemoveAll(tempDir); err1 != nil {
			logger.Warn("failed to remove temp dir", zap.String("temp_dir", tempDir), zap.Error(err1))
		}
	}()
	var tempFile *os.File
	tempFile, err = os.CreateTemp(tempDir, "")
	if err != nil {
		return errors.Annotate(err, "create temp file")
	}
	defer func() {
		if err1 := tempFile.Close(); err1 != nil {
			logger.Warn("failed to close temp file", zap.String("temp_file", tempFile.Name()), zap.Error(err1))
		}
	}()
	var fileSize int64
	err = parser.Register("file", func(r io.Reader, header formstream.Header) error {
		fileSize, err = io.Copy(tempFile, r)
		if err != nil {
			return errors.Annotate(err, "copy body to tmp file")
		}
		return nil
	})
	if err != nil {
		return errors.Annotate(err, "register form")
	}
	err = parser.Parse(r.Body)
	if err != nil {
		return errors.Annotate(err, "parse form")
	}
	var zipReader *zip.Reader
	zipReader, err = zip.NewReader(tempFile, fileSize)
	if err != nil {
		return errors.Annotate(err, "new zip reader")
	}
	info := &DiffApplyInfo{
		Num: len(zipReader.File),
	}
	for _, file := range zipReader.File {
		info.TotalSize += file.FileInfo().Size()
	}
	if err = watcher.SendStatus(info); err != nil {
		return errors.Annotate(err, "send status")
	}
	var zipFile io.ReadCloser
	for _, file := range zipReader.File {
		zipFile, err = file.Open()
		if err != nil {
			return errors.Annotate(err, "open zip file")
		}
		err = g.Cacher.Put(r.Context(), file.Name, file.FileInfo().Size(), watcher.WatchReader(zipFile))
		if err != nil {
			return errors.Annotate(err, "put zip file")
		}
		err = watcher.AddDir(1)
		if err != nil {
			return errors.Annotate(err, "add dir")
		}
	}
	return nil
}
