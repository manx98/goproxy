/*
Package goproxy implements a minimalist Go module proxy handler.
*/
package goproxy

import (
	"context"
	"errors"
	"fmt"
	"github.com/goproxy/goproxy/cache"
	"github.com/goproxy/goproxy/db"
	"github.com/goproxy/goproxy/logger"
	"go.uber.org/zap"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	"golang.org/x/mod/module"
)

// tempDirPattern is the pattern for creating temporary directories.
const tempDirPattern = "goproxy.tmp.*"

// Goproxy is the top-level struct of this project.
//
// For requests involving the download of a large number of modules (e.g., for
// bulk static analysis), Goproxy supports a non-standard header,
// "Disable-Module-Fetch: true", which instructs it to return only cached
// content.
type Goproxy struct {
	// Fetcher is used to fetch module files.
	//
	// If Fetcher is nil, [GoFetcher] is used.
	//
	// Note that any error returned by Fetcher that matches [fs.ErrNotExist]
	// will result in a 404 response with the error message in the response
	// body.
	Fetcher Fetcher

	// ProxiedSumDBs is a list of proxied checksum databases (see
	// https://go.dev/design/25530-sumdb#proxying-a-checksum-database). Each
	// entry is in the form "<sumdb-name>" or "<sumdb-name> <sumdb-URL>".
	// The first form is a shorthand for the second, where the corresponding
	// <sumdb-URL> will be the <sumdb-name> itself as a host with an "https"
	// scheme. Invalid entries will be silently ignored.
	//
	// If ProxiedSumDBs contains duplicate checksum database names, only the
	// last value in the slice for each duplicate checksum database name is
	// used.
	ProxiedSumDBs []string

	// Cacher is used to cache module files.
	//
	// If Cacher is nil, caching is disabled.
	Cacher cache.Cacher

	// TempDir is the directory for storing temporary files.
	//
	// If TempDir is empty, [os.TempDir] is used.
	TempDir string

	// Transport is used to execute outgoing requests.
	//
	// If Transport is nil, [http.DefaultTransport] is used.
	Transport http.RoundTripper

	NoFetch bool

	initOnce      sync.Once
	fetcher       Fetcher
	proxiedSumDBs map[string]*url.URL
	httpClient    *http.Client
}

// init initializes the g.
func (g *Goproxy) init() {
	g.fetcher = g.Fetcher
	if g.fetcher == nil {
		g.fetcher = &GoFetcher{TempDir: g.TempDir, Transport: g.Transport}
	}

	g.proxiedSumDBs = map[string]*url.URL{}
	for _, sumdb := range g.ProxiedSumDBs {
		parts := strings.Fields(sumdb)
		if len(parts) == 0 {
			continue
		}
		name := parts[0]
		rawURL := "https://" + name
		if len(parts) > 1 {
			rawURL = parts[1]
		}
		u, err := url.Parse(rawURL)
		if err != nil {
			continue
		}
		g.proxiedSumDBs[name] = u
	}

	g.httpClient = &http.Client{Transport: g.Transport}
}

// ServeHTTP implements [http.Handler].
func (g *Goproxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	g.initOnce.Do(g.init)
	logger.Debug("accept request", zap.String("url", req.URL.String()))
	switch req.Method {
	case http.MethodGet, http.MethodHead:
	default:
		responseMethodNotAllowed(rw, req, 86400)
		return
	}

	urlPath := cleanPath(req.URL.Path)
	if urlPath != req.URL.Path || urlPath[len(urlPath)-1] == '/' {
		responseNotFound(rw, req, 86400)
		return
	}
	target := urlPath[1:] // Remove the leading slash.

	if strings.HasPrefix(target, "sumdb/") {
		g.serveSumDB(rw, req, target)
		return
	}
	g.serveFetch(rw, req, target)
}

// serveFetch serves fetch requests.
func (g *Goproxy) serveFetch(rw http.ResponseWriter, req *http.Request, target string) {
	if !db.Lock.TryRLock() {
		logger.Warn("The current request will be rejected because the export operation is running.", zap.String("method", req.Method), zap.String("URL", req.URL.String()))
		responseInternalServerError(rw, req)
		return
	}
	defer db.Lock.RUnlock()
	escapedModulePath, after, ok := strings.Cut(target, "/@")
	if !ok {
		responseNotFound(rw, req, 86400, "missing /@v/")
		return
	}
	modulePath, err := module.UnescapePath(escapedModulePath)
	if err != nil {
		responseNotFound(rw, req, 86400, err)
		return
	}
	switch after {
	case "latest":
		g.serveFetchQuery(rw, req, target, modulePath, after)
		return
	case "v/list":
		g.serveFetchList(rw, req, target, modulePath)
		return
	}

	if !strings.HasPrefix(after, "v/") {
		responseNotFound(rw, req, 86400, "missing /@v/")
		return
	}
	after = after[2:] // Remove the leading "v/".
	ext := path.Ext(after)
	switch ext {
	case ".info", ".mod", ".zip":
	case "":
		responseNotFound(rw, req, 86400, fmt.Sprintf("no file extension in filename %q", after))
		return
	default:
		responseNotFound(rw, req, 86400, fmt.Sprintf("unexpected extension %q", ext))
		return
	}

	escapedModuleVersion := strings.TrimSuffix(after, ext)
	moduleVersion, err := module.UnescapeVersion(escapedModuleVersion)
	if err != nil {
		responseNotFound(rw, req, 86400, err)
		return
	}
	switch moduleVersion {
	case "latest", "upgrade", "patch":
		responseNotFound(rw, req, 86400, "invalid version")
		return
	}
	if checkCanonicalVersion(modulePath, moduleVersion) == nil {
		g.serveFetchDownload(rw, req, target, modulePath, moduleVersion)
	} else if ext == ".info" {
		g.serveFetchQuery(rw, req, target, modulePath, moduleVersion)
	} else {
		responseNotFound(rw, req, 86400, "unrecognized version")
	}
}

// serveFetchQuery serves fetch query requests.
func (g *Goproxy) serveFetchQuery(rw http.ResponseWriter, req *http.Request, target, modulePath, moduleQuery string) {
	const (
		contentType        = "application/json; charset=utf-8"
		cacheControlMaxAge = 60
	)
	if g.NoFetch {
		g.serveCache(rw, req, target, contentType, cacheControlMaxAge, nil)
		return
	}
	version, time, err := g.fetcher.Query(req.Context(), modulePath, moduleQuery)
	if err != nil {
		g.serveCache(rw, req, target, contentType, cacheControlMaxAge, func() {
			logger.Error("failed to query module version", zap.String("target", target), zap.Error(err))
			responseError(rw, req, err, true)
		})
		return
	}
	g.servePutCache(rw, req, target, contentType, cacheControlMaxAge, strings.NewReader(marshalInfo(version, time)))
}

// serveFetchList serves fetch list requests.
func (g *Goproxy) serveFetchList(rw http.ResponseWriter, req *http.Request, target, modulePath string) {
	const (
		contentType        = "text/plain; charset=utf-8"
		cacheControlMaxAge = 60
	)
	if g.NoFetch {
		g.serveCache(rw, req, target, contentType, cacheControlMaxAge, nil)
		return
	}
	versions, err := g.fetcher.List(req.Context(), modulePath)
	if err != nil {
		g.serveCache(rw, req, target, contentType, cacheControlMaxAge, func() {
			logger.Error("failed to list module versions", zap.String("target", target), zap.Error(err))
			responseError(rw, req, err, true)
		})
		return
	}
	g.servePutCache(rw, req, target, contentType, cacheControlMaxAge, strings.NewReader(strings.Join(versions, "\n")))
}

// serveFetchDownload serves fetch download requests.
func (g *Goproxy) serveFetchDownload(rw http.ResponseWriter, req *http.Request, target, modulePath, moduleVersion string) {
	const cacheControlMaxAge = 604800

	ext := path.Ext(target)
	var contentType string
	switch ext {
	case ".info":
		contentType = "application/json; charset=utf-8"
	case ".mod":
		contentType = "text/plain; charset=utf-8"
	case ".zip":
		contentType = "application/zip"
	}

	if g.NoFetch {
		g.serveCache(rw, req, target, contentType, cacheControlMaxAge, nil)
		return
	}

	if content, err := g.cache(req.Context(), target); err == nil {
		responseSuccess(rw, req, content, contentType, cacheControlMaxAge)
		return
	} else if !errors.Is(err, fs.ErrNotExist) {
		logger.Error("failed to get cached module file", zap.String("target", target), zap.Error(err))
		responseInternalServerError(rw, req)
		return
	}

	info, mod, zip, err := g.fetcher.Download(req.Context(), modulePath, moduleVersion)
	if err != nil {
		logger.Error("failed to download module version", zap.String("target", target), zap.Error(err))
		responseError(rw, req, err, false)
		return
	}
	defer func() {
		info.Close()
		mod.Close()
		zip.Close()
	}()

	targetWithoutExt := strings.TrimSuffix(target, path.Ext(target))
	for _, ch := range []struct {
		ext     string
		content io.ReadSeeker
	}{
		{".info", info},
		{".mod", mod},
		{".zip", zip},
	} {
		if err := g.putCache(req.Context(), targetWithoutExt+ch.ext, ch.content); err != nil {
			logger.Error("failed to ch module file", zap.String("target", target), zap.Error(err))
			responseInternalServerError(rw, req)
			return
		}
	}

	var content io.ReadSeeker
	switch ext {
	case ".info":
		content = info
	case ".mod":
		content = mod
	case ".zip":
		content = zip
	}
	if _, err := content.Seek(0, io.SeekStart); err != nil {
		logger.Error("failed to seek", zap.Error(err))
		responseInternalServerError(rw, req)
		return
	}
	responseSuccess(rw, req, content, contentType, 604800)
}

// serveSumDB serves checksum database proxy requests.
func (g *Goproxy) serveSumDB(rw http.ResponseWriter, req *http.Request, target string) {
	name, path, ok := strings.Cut(strings.TrimPrefix(target, "sumdb/"), "/")
	if !ok {
		responseNotFound(rw, req, 86400)
		return
	}
	path = "/" + path // Add the leading slash back.
	u, ok := g.proxiedSumDBs[name]
	if !ok {
		responseNotFound(rw, req, 86400)
		return
	}

	var (
		contentType        string
		cacheControlMaxAge int
	)
	switch {
	case path == "/supported":
		setResponseCacheControlHeader(rw, 86400)
		rw.WriteHeader(http.StatusOK)
		return
	case path == "/latest":
		contentType = "text/plain; charset=utf-8"
		cacheControlMaxAge = 3600
	case strings.HasPrefix(path, "/lookup/"):
		contentType = "text/plain; charset=utf-8"
		cacheControlMaxAge = 86400
	case strings.HasPrefix(path, "/tile/"):
		contentType = "application/octet-stream"
		cacheControlMaxAge = 86400
	default:
		responseNotFound(rw, req, 86400)
		return
	}
	if g.NoFetch {
		g.serveCache(rw, req, target, contentType, cacheControlMaxAge, nil)
		return
	}
	tempDir, err := os.MkdirTemp(g.TempDir, tempDirPattern)
	if err != nil {
		logger.Error("failed to create temporary directory", zap.Error(err))
		responseInternalServerError(rw, req)
		return
	}
	defer os.RemoveAll(tempDir)

	file, err := httpGetTemp(req.Context(), g.httpClient, appendURL(u, path).String(), tempDir)
	if err != nil {
		g.serveCache(rw, req, target, contentType, cacheControlMaxAge, func() {
			logger.Error("failed to proxy checksum database", zap.String("target", target), zap.Error(err))
			responseError(rw, req, err, true)
		})
		return
	}
	g.servePutCacheFile(rw, req, target, contentType, cacheControlMaxAge, file)
}

// serveCache serves requests with cached module files.
func (g *Goproxy) serveCache(rw http.ResponseWriter, req *http.Request, name, contentType string, cacheControlMaxAge int, onNotFound func()) {
	content, err := g.cache(req.Context(), name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if onNotFound != nil {
				onNotFound()
			} else {
				responseNotFound(rw, req, 60, "temporarily unavailable")
			}
			return
		}
		logger.Error("failed to get cached module file", zap.String("name", name), zap.Error(err))
		responseInternalServerError(rw, req)
		return
	}
	defer content.Close()
	responseSuccess(rw, req, content, contentType, cacheControlMaxAge)
}

// servePutCache serves requests after putting the content to the g.Cacher.
func (g *Goproxy) servePutCache(rw http.ResponseWriter, req *http.Request, name, contentType string, cacheControlMaxAge int, content io.ReadSeeker) {
	if err := g.putCache(req.Context(), name, content); err != nil {
		logger.Error("failed to cache module file", zap.String("name", name), zap.Error(err))
		responseInternalServerError(rw, req)
		return
	}
	if _, err := content.Seek(0, io.SeekStart); err != nil {
		logger.Error("failed to seek", zap.Error(err))
		responseInternalServerError(rw, req)
		return
	}
	responseSuccess(rw, req, content, contentType, cacheControlMaxAge)
}

// servePutCacheFile is like [servePutCache] but reads the content from the
// local file.
func (g *Goproxy) servePutCacheFile(rw http.ResponseWriter, req *http.Request, name, contentType string, cacheControlMaxAge int, file string) {
	f, err := os.Open(file)
	if err != nil {
		logger.Error("failed to open file", zap.Error(err))
		responseInternalServerError(rw, req)
		return
	}
	defer f.Close()
	g.servePutCache(rw, req, name, contentType, cacheControlMaxAge, f)
}

// cache returns the matched cache for the name from the g.Cacher.
func (g *Goproxy) cache(ctx context.Context, name string) (io.ReadCloser, error) {
	if g.Cacher == nil {
		return nil, fs.ErrNotExist
	}
	return g.Cacher.Get(ctx, name)
}

// putCache puts a cache to the g.Cacher for the name with the content.
func (g *Goproxy) putCache(ctx context.Context, name string, content io.ReadSeeker) error {
	if g.Cacher == nil {
		return nil
	}
	return g.Cacher.Put(ctx, name, content)
}

// putCacheFile is like [putCache] but reads the content from the local file.
func (g *Goproxy) putCacheFile(ctx context.Context, name, file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return g.putCache(ctx, name, f)
}

// cleanPath returns the canonical path for the p.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	if p[len(p)-1] == '/' && np != "/" {
		if len(p) == len(np)+1 && strings.HasPrefix(p, np) {
			return p
		}
		return np + "/"
	}
	return np
}
