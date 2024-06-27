package web

import (
	"github.com/goproxy/goproxy/constant"
	"github.com/goproxy/goproxy/export"
	"github.com/goproxy/goproxy/logger"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"golang.org/x/mod/modfile"
	"io"
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

func (g *GoTool) Get(w http.ResponseWriter, modFile io.Reader) (err error) {
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
			logger.Warn("failed to send close with error", zap.NamedError("occur_error", err))
			if err1 := st.Close(err.Error()); err1 != nil {
				logger.Warn("failed to send close with error", zap.NamedError("occur_error", err), zap.Error(err))
			}
		} else {
			if err = st.Close(""); err != nil {
				logger.Warn("failed to send close with success", zap.Error(err))
			}
		}
	}()
	modData, err := io.ReadAll(modFile)
	if err != nil {
		return errors.Annotate(err, "failed to read mod file")
	}
	modStruct, err := modfile.Parse("go.mod", modData, nil)
	if err != nil {
		return errors.Annotate(err, "failed to parse mod file")
	}
	var tempDir string
	err = os.MkdirAll(g.TmpPath, 0o755)
	if err != nil {
		return errors.Annotate(err, "failed to create mod get tmp dir")
	}
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
	command := exec.Command(g.GoBin, "mod", "init", "mod_get")
	command.Dir = modDir
	command.Stderr = stdErr
	command.Stdout = stdOut
	command.Env = env
	if err = command.Run(); err != nil {
		return errors.Annotate(err, "failed to init mod")
	}
	for _, req := range modStruct.Require {
		if req.Indirect {
			continue
		}
		command = exec.Command(g.GoBin, "get", "-x", req.Mod.String())
		command.Dir = modDir
		command.Stderr = stdErr
		command.Stdout = stdOut
		command.Env = env
		if err = command.Run(); err != nil {
			return errors.Annotate(err, "failed to download mod")
		}
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
		upValue := strings.ToUpper(value)
		if strings.HasPrefix(upValue, "GOPROXY=") {
			foundGoProxy = true
			t.Env = append(t.Env, goProxy)
		} else if strings.HasPrefix(upValue, "GOSUMDB=") {
			t.Env = append(t.Env, goSumDb)
			foundGoSumDb = true
		} else if !strings.HasPrefix(upValue, "GOPATH=") {
			t.Env = append(t.Env, value)
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
