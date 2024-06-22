package web

import (
	"github.com/goproxy/goproxy"
	"net/http"
)

func Init(mux *http.ServeMux, p *goproxy.Goproxy) {
	initApi(mux, p)
}
