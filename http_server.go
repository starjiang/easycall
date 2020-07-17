package easycall

import (
	"net/http"
	"strconv"
	"time"

	"github.com/starjiang/elog"
)

const (
	HTTP_READ_WRITE_TIMEOUT = 10
)

//Server for EasyService
type HttpServer struct {
}

func (serv *HttpServer) CreateHttpServer(port int, handler http.Handler) error {

	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(port),
		Handler:      handler,
		ReadTimeout:  HTTP_READ_WRITE_TIMEOUT * time.Second,
		WriteTimeout: HTTP_READ_WRITE_TIMEOUT * time.Second,
	}

	err := srv.ListenAndServe()
	if err != nil {
		elog.Error("ListenAndServe: ", err)
		return err
	}
	return nil
}
