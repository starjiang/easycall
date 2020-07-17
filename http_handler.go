package easycall

import (
	"net/http"

	"github.com/starjiang/elog"
)

type HttpHandler struct {
	middlewares []*HttpMiddlewareInfo
}

func NewHttpHandler(middlewares []*HttpMiddlewareInfo) *HttpHandler {
	return &HttpHandler{middlewares: middlewares}
}

type HttpMiddlewareFunc func(w http.ResponseWriter, r *http.Request, next *HttpMiddlewareInfo)

type HttpMiddlewareInfo struct {
	Middleware HttpMiddlewareFunc
	Next       *HttpMiddlewareInfo
}

func (hh *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if len(hh.middlewares) > 0 {
		hh.middlewares[0].Middleware(w, r, hh.middlewares[0].Next)
	} else {
		elog.Error("HttpMiddleware chain is empty")
	}
}

func (hh *HttpHandler) AddMiddleware(middleware HttpMiddlewareFunc) *HttpHandler {
	if hh.middlewares == nil {
		hh.middlewares = make([]*HttpMiddlewareInfo, 0)
	}
	minfo := &HttpMiddlewareInfo{middleware, nil}

	mlen := len(hh.middlewares)
	if mlen > 0 {
		hh.middlewares[mlen-1].Next = minfo
	}

	hh.middlewares = append(hh.middlewares, minfo)
	return hh
}
