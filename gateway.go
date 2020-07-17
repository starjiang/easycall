package easycall

import "net/http"

type Gateway struct {
}

func (gw *Gateway) StartGWServer(port int, handler *GatewayHandler) error {
	server := &Server{}
	return server.CreateServer(port, handler)
}

func (gw *Gateway) StartHttpGWServer(port int, handler http.Handler) error {
	httpServer := &HttpServer{}
	return httpServer.CreateHttpServer(port, handler)
}
