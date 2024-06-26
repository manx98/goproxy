package web

import (
	"crypto/md5"
	"embed"
	"encoding/hex"
	"fmt"
	"github.com/goproxy/goproxy"
	"github.com/goproxy/goproxy/utils"
	"go.uber.org/zap"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

func Init(mux *http.ServeMux, p *goproxy.Goproxy) {
	InitWebResource(mux)
	initApi(mux, p)
}

//go:embed dist/assets
var sourceFs embed.FS

//go:embed dist/index.html
var indexHtml string

var IndexHandler func(ctx http.ResponseWriter, r *http.Request)

func InitWebResource(router *http.ServeMux) {
	err := fs.WalkDir(sourceFs, ".", func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			open, err := sourceFs.Open(path)
			if err != nil {
				zap.L().Fatal("无法打开内嵌文件", zap.String("path", path), zap.Error(err))
			}
			fileContent, err := io.ReadAll(open)
			if err != nil {
				zap.L().Fatal("无法读取内嵌文件", zap.String("path", path), zap.Error(err))
			}
			router.HandleFunc(strings.TrimLeft(path, "dist"), warpStaticRoute(fileContent, path))
		}
		return err
	})
	if err != nil {
		zap.L().Fatal("初始化静态资源列表出现异常", zap.Error(err))
	}
	IndexHandler = warpStaticRoute([]byte(indexHtml), "index.html")
}

func warpStaticRoute(content []byte, file string) func(ctx http.ResponseWriter, r *http.Request) {
	sum := md5.Sum(content)
	contentType := mime.TypeByExtension(filepath.Ext(file))
	contentMd5 := hex.EncodeToString(sum[:])
	return func(ctx http.ResponseWriter, r *http.Request) {
		header := ctx.Header()
		ifNoneMatch := header.Get("If-None-Match")
		if ifNoneMatch == contentMd5 {
			ctx.WriteHeader(http.StatusNotModified)
			return
		}
		header.Set("Cache-Control", "no-cache")
		header.Set("ETag", contentMd5)
		header.Set("Content-Type", contentType)
		header.Set("Accept-Ranges", "bytes")
		var err error
		var start int64
		var end int64
		if len(content) == 0 {
			ctx.WriteHeader(http.StatusOK)
			return
		} else {
			start, end, err = utils.HandleRange(header.Get("Range"))
			if err != nil {
				start = 0
				end = int64(len(content) - 1)
			} else {
				if start < 0 || start >= int64(len(content)) {
					start = 0
				}
				if end < 0 || end >= int64(len(content)) {
					end = int64(len(content) - 1)
				}
			}
			header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(content)))
			header.Set("Content-Length", strconv.FormatInt(end-start+1, 10))
			if end-start+1 == int64(len(content)) {
				ctx.WriteHeader(http.StatusOK)
			} else {
				ctx.WriteHeader(http.StatusPartialContent)
			}
			_, err = ctx.Write(content[start : end+1])
			if err != nil {
				zap.L().Debug("failed to write content")
			}
		}
	}
}
