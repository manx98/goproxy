package web

import (
	"github.com/goproxy/goproxy/constant"
	"github.com/goproxy/goproxy/export"
	"github.com/goproxy/goproxy/logger"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

func (g *GoTool) Get(w http.ResponseWriter, name string, version string) (err error) {
	st := export.NewCreateCheckPointWatcher(w)
	stdErr := WriterFunc(func(p []byte) (n int, err error) {
		defer func() {
			n -= 1
		}()
		return st.Write(append(p, constant.Stderr))
	})
	stdOut := WriterFunc(func(p []byte) (n int, err error) {
		defer func() {
			n -= 1
		}()
		return st.Write(append(p, constant.Stdout))
	})
	defer func() {
		if err != nil {
			if err1 := st.Close(err.Error()); err1 != nil {
				logger.Warn("failed to send close with error", zap.NamedError("occur_error", err), zap.Error(err))
			}
		} else {
			if err = st.Close(""); err != nil {
				logger.Warn("failed to send close with success", zap.Error(err))
			}
		}
	}()
	var tempDir string
	tempDir, err = os.MkdirTemp(g.TmpPath, constant.TempDirModeDownloadPattern)
	if err != nil {
		return errors.Annotate(err, "failed to create mod get tmp dir")
	}
	defer func() {
		if err1 := os.RemoveAll(tempDir); err1 != nil {
			logger.Warn("failed to remove mod get tmp dir", zap.Error(err1), zap.String("path", tempDir))
		}
	}()
	goPath := filepath.Join(tempDir, "gopath")
	if err = os.Mkdir(goPath, 0755); err != nil {
		return errors.Annotate(err, "failed to create mod get tmp dir")
	}
	modDir := filepath.Join(tempDir, "mod")
	if err = os.Mkdir(modDir, 0755); err != nil {
		return errors.Annotate(err, "failed to create mod get tmp dir")
	}
	env := append([]string{"GOPATH=" + goPath}, g.Env...)
	command := exec.Command(g.GoBin, "mod", "init", "go_get")
	command.Env = env
	command.Dir = modDir
	command.Stderr = stdErr
	command.Stdout = stdOut
	if err = command.Run(); err != nil {
		return errors.Annotate(err, "failed to init mod get tmp dir")
	}
	command = exec.Command(g.GoBin, "get", "-x", name+"@"+version)
	command.Dir = modDir
	command.Stderr = stdErr
	command.Stdout = stdOut
	command.Env = env
	if err = command.Run(); err != nil {
		return errors.Annotate(err, "failed to download mod")
	}
	return
}

func NewGoTool(goBin string, goProxy string, tmpDir string) *GoTool {
	if goBin == "" {
		goBin = "go"
	}
	t := &GoTool{
		GoBin:   goBin,
		TmpPath: tmpDir,
	}
	environ := os.Environ()
	foundGoProxy := false
	foundGoSumDb := false
	goProxy = "GOPROXY=" + goProxy + ",direct"
	goSumDb := "GOSUMDB=off"
	for _, value := range environ {
		value = strings.ToUpper(value)
		if strings.HasPrefix(value, "GOPROXY=") {
			foundGoProxy = true
			t.Env = append(t.Env, goProxy)
		} else if strings.HasPrefix(value, "GOSUMDB=") {
			t.Env = append(t.Env, goSumDb)
			foundGoSumDb = true
		} else if !strings.HasPrefix(value, "GOPATH=") {

		}
	}
	if !foundGoProxy {
		t.Env = append(t.Env, goProxy)
	}
	if !foundGoSumDb {
		t.Env = append(t.Env, goSumDb)
	}
	return t
}
