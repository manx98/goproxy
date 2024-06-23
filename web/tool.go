package web

import (
	"github.com/goproxy/goproxy/constant"
	"github.com/goproxy/goproxy/export"
	"github.com/goproxy/goproxy/logger"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type WriterFunc func(p []byte) (n int, err error)

func (w WriterFunc) Write(p []byte) (n int, err error) {
	return w(p)
}

type GoTool struct {
	GoBin   string
	Env     []string
	TmpPath string
}

func (g *GoTool) Get(w http.ResponseWriter, query string) {
	command := exec.Command(g.GoBin, "get", "-x", query)
	command.Dir = g.TmpPath
	st := export.NewCreateCheckPointWatcher(w)
	command.Stderr = WriterFunc(func(p []byte) (n int, err error) {
		defer func() {
			n -= 1
		}()
		return st.Write(append(p, constant.Stderr))
	})
	command.Stdout = WriterFunc(func(p []byte) (n int, err error) {
		defer func() {
			n -= 1
		}()
		return st.Write(append(p, constant.Stdout))
	})
	command.Env = g.Env
	if err := command.Run(); err != nil {
		if err1 := st.Close(err.Error()); err1 != nil {
			logger.Warn("failed to send close with error", zap.NamedError("occur_error", err), zap.Error(err))
		}
	} else {
		if err = st.Close(""); err != nil {
			logger.Warn("failed to send close with success", zap.Error(err))
		}
	}
}

func NewGoTool(goBin string, goProxy string, goPath string) *GoTool {
	if goBin == "" {
		goBin = "go"
	}
	t := &GoTool{
		GoBin:   goBin,
		TmpPath: goPath,
	}
	environ := os.Environ()
	foundGoProxy := false
	foundGoPath := false
	foundGoSumDb := false
	goProxy = "GOPROXY=" + goProxy + ",direct"
	goPath = "GOPATH=" + goPath
	goSumDb := "GOSUMDB=off"
	for i, value := range environ {
		if strings.HasPrefix(value, "GOPROXY=") {
			environ[i] = goProxy
			foundGoProxy = true
		} else if strings.HasPrefix(value, "GOPATH=") {
			environ[i] = goPath
			foundGoPath = true
		} else if strings.HasPrefix(value, "GOSUMDB=") {
			environ[i] = goSumDb
			foundGoSumDb = true
		}
	}
	if !foundGoProxy {
		environ = append(environ, goProxy)
	}
	if !foundGoPath {
		environ = append(environ, goPath)
	}
	if !foundGoSumDb {
		environ = append(environ, goSumDb)
	}
	t.Env = environ
	return t
}
