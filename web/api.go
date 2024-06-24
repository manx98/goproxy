package web

import (
	"encoding/base64"
	"github.com/goproxy/goproxy"
	"github.com/goproxy/goproxy/constant"
	"github.com/goproxy/goproxy/db"
	"github.com/goproxy/goproxy/export"
	"github.com/goproxy/goproxy/logger"
	"github.com/goproxy/goproxy/obj"
	"github.com/juju/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

func initApi(mux *http.ServeMux, p *goproxy.Goproxy) {
	mux.HandleFunc("/api/get_header", getHeader)
	mux.HandleFunc("/api/download_diff", func(writer http.ResponseWriter, request *http.Request) {
		downloadDiff(writer, request, p)
	})
	mux.HandleFunc("/api/create_checkpoint", func(writer http.ResponseWriter, request *http.Request) {
		createCheckpoint(writer, request, p)
	})
	tool := NewGoTool(p.GoBin, p.Address, p.TempDir)
	mux.HandleFunc("/api/download_mod", func(writer http.ResponseWriter, request *http.Request) {
		downloadMod(writer, request, tool)
	})
}

func getHeader(w http.ResponseWriter, r *http.Request) {
	requestId, requestIdLog := nextRequestId()
	query := r.URL.Query()
	idStr := query.Get("id")
	numStr := query.Get("num")
	logger.Debug("start get header", zap.String("id", idStr), zap.String("num", numStr))
	var num int
	var err error
	if numStr == "" {
		num = 10
	} else {
		num, err = strconv.Atoi(numStr)
		if err != nil {
			logger.Debug("failed to parse num", zap.Error(err), requestIdLog)
			responseBadArgs(requestId, w)
			return
		}
	}
	var id []byte
	if idStr != "" {
		id, err = base64.StdEncoding.DecodeString(idStr)
		if err != nil {
			logger.Debug("invalid id str", zap.String("id", idStr), zap.Error(err), requestIdLog)
			responseBadArgs(requestId, w)
			return
		}
	}
	var result []*obj.CheckPoint
	err = db.View(func(tx *bbolt.Tx) error {
		var first *obj.CheckPoint
		if len(id) == 0 {
			first, err = export.GetHead(tx)
			if err != nil {
				return err
			}
			if first == nil {
				return nil
			}
			result = append(result, first)
			id = first.Parent
		}
		for len(result) < num {
			first, err = export.GetCheckPoint(tx, id)
			if err != nil {
				return err
			}
			if first == nil {
				return nil
			}
			result = append(result, first)
			id = first.Parent
		}
		return nil
	})
	if err != nil {
		logger.Warn("failed to get header", requestIdLog, zap.Error(err))
		responseInternalError(requestId, w, "load check point occur error")
	} else {
		responseSuccess(requestId, w, result)
	}
}

func downloadDiff(w http.ResponseWriter, r *http.Request, g *goproxy.Goproxy) {
	requestId, requestIdLog := nextRequestId()
	query := r.URL.Query()
	idStr := query.Get("id")
	logger.Debug("start download diff", zap.String("id", idStr), requestIdLog)
	id := constant.EmptyId
	st := export.NewCreateCheckPointWatcher(w)
	err := db.View(func(tx *bbolt.Tx) (err error) {
		if idStr == "" {
			var head *obj.CheckPoint
			head, err = export.GetHead(tx)
			if head != nil {
				id = head.Id
			}
			if err != nil {
				logger.Warn("failed to get head", requestIdLog, zap.Error(err))
				responseInternalError(requestId, w, "load last head")
				return nil
			}
		} else {
			id, err = base64.StdEncoding.DecodeString(idStr)
			if err != nil {
				logger.Debug("invalid id str", zap.String("id", idStr), zap.Error(err), requestIdLog)
				responseBadArgs(requestId, w)
				return nil
			}
		}
		zipper := export.NewDiffZipper(g.Cacher, w, st)
		defer zipper.Close()
		return zipper.DiffHead(r.Context(), tx, id)
	})
	if err != nil {
		if err1 := st.Close(err.Error()); err1 != nil {
			logger.Warn("failed to finish download diff with error", requestIdLog, zap.NamedError("occur_error", err), zap.Error(err1))
		}
	} else {
		if err = st.Close(""); err != nil {
			logger.Warn("failed to finish download diff", requestIdLog, zap.Error(err))
		}
	}
}

func createCheckpoint(w http.ResponseWriter, r *http.Request, g *goproxy.Goproxy) {
	st := export.NewCreateCheckPointWatcher(w)
	var id []byte
	err := db.Update(func(tx *bbolt.Tx) (err error) {
		id, err = export.CreateCheckPoint(tx, r.Context(), r.URL.Query().Get("desc"), g.Cacher, st)
		return
	})
	if err == nil {
		if _, err = st.Write(id); err != nil {
			err = errors.Annotate(err, "write head id")
		}
	}
	if err != nil {
		logger.Warn("failed to create checkpoint", zap.Error(err))
		if err1 := st.Close(err.Error()); err1 != nil {
			logger.Warn("failed to finish create checkpoint with error", zap.NamedError("occur_error", err), zap.Error(err1))
		}
	} else {
		if err = st.Close(""); err != nil {
			logger.Warn("failed to finish create checkpoint", zap.Error(err))
		}
	}
}

func downloadMod(w http.ResponseWriter, r *http.Request, f *GoTool) {
	query := r.URL.Query()
	name := query.Get("q")
	version := query.Get("v")
	f.Get(w, name, version)
}
