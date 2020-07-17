package easycall

import (
	"github.com/panjf2000/ants/v2"
	"github.com/starjiang/elog"
)

type GWMiddlewareFunc func(reqPkg *EasyPackage, client *EasyConnection, next *GWMiddlewareInfo)

type GWMiddlewareInfo struct {
	Middleware GWMiddlewareFunc
	Next       *GWMiddlewareInfo
}

//GatewayHandler for Gateway
type GatewayHandler struct {
	middlewares []*GWMiddlewareInfo
	pool        *ants.Pool
}

func NewGatewayHandler(middlewares []*GWMiddlewareInfo) *GatewayHandler {
	gatewayHandler := &GatewayHandler{}
	pool, _ := ants.NewPool(EASYCALL_SERVICE_GO_POOL_SIZE, ants.WithNonblocking(true))
	gatewayHandler.pool = pool
	gatewayHandler.middlewares = middlewares
	return gatewayHandler
}

func (h *GatewayHandler) AddMiddleware(middleware GWMiddlewareFunc) *GatewayHandler {

	if h.middlewares == nil {
		h.middlewares = make([]*GWMiddlewareInfo, 0)
	}
	minfo := &GWMiddlewareInfo{middleware, nil}

	mlen := len(h.middlewares)
	if mlen > 0 {
		h.middlewares[mlen-1].Next = minfo
	}

	h.middlewares = append(h.middlewares, minfo)
	return h
}

func (h *GatewayHandler) Dispatch(pkgData []byte, client *EasyConnection) {

	err := h.pool.Submit(func() {

		defer PanicHandler()

		reqPkg, err := DecodeWithBodyData(pkgData)

		if err != nil {
			elog.Error("decode pkg fail:", err)
			return
		}
		if len(h.middlewares) > 0 {
			h.middlewares[0].Middleware(reqPkg, client, h.middlewares[0].Next)
		} else {
			elog.Error("GWMiddleware chain is empty")
		}

	})

	if err != nil {
		elog.Error("submit to pool fail,", err)
	}
}
