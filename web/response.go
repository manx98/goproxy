package web

import (
	"encoding/json"
	"github.com/bwmarrin/snowflake"
	"github.com/goproxy/goproxy/logger"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"net/http"
)

var snowflakeNode *snowflake.Node

func init() {
	var err error
	snowflakeNode, err = snowflake.NewNode(0)
	if err != nil {
		logger.Fatal("failed to init snowflake", zap.Error(err))
	}
}

func requestIdToLogField(requestId int64) zap.Field {
	return zap.Int64("request_id", requestId)
}

func nextRequestId() (requestId int64, requestIdLog zap.Field) {
	requestId = snowflakeNode.Generate().Int64()
	requestIdLog = requestIdToLogField(requestId)
	return
}

type ResponseData struct {
	Code      int         `json:"code,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Msg       string      `json:"msg,omitempty"`
	RequestId int64       `json:"request_id"`
}

func responseWriteJson(w http.ResponseWriter, code int, data *ResponseData) (err error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return errors.Annotatef(err, "marshal json data")
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(dataBytes)
	return errors.Annotate(err, "write data")
}

func responseBadArgs(reqId int64, w http.ResponseWriter) {
	err := responseWriteJson(w, 200, &ResponseData{
		Code:      400,
		Msg:       "bad args",
		RequestId: reqId,
	})
	if err != nil {
		logger.Debug("failed to write data", zap.Error(err))
	}
}

func responseInternalError(reqId int64, w http.ResponseWriter, msg string) {
	err := responseWriteJson(w, 200, &ResponseData{
		Code:      500,
		Msg:       msg,
		RequestId: reqId,
	})
	if err != nil {
		logger.Debug("failed to write data", zap.Error(err))
	}
}

func responseSuccess(reqId int64, w http.ResponseWriter, data interface{}) {
	err := responseWriteJson(w, 200, &ResponseData{
		Code:      0,
		Data:      data,
		RequestId: reqId,
	})
	if err != nil {
		logger.Debug("failed to write data", zap.Error(err), requestIdToLogField(reqId))
	}
}
