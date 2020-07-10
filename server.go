package easycall

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/starjiang/elog"
)

const (
	TCP_KEEPALIVE_PERIOD = 15
)

//Server for EasyService
type Server struct {
}

func (serv *Server) CreateServer(port int, service interface{}, middlewares []*MiddlewareInfo) error {

	handler := NewServiceHandler(service, middlewares)
	listen, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		elog.Error("listen error: ", err)
		return err
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			elog.Error("accept error: ", err)
			continue
		}
		tcpConn := conn.(*net.TCPConn)
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(time.Minute * TCP_KEEPALIVE_PERIOD)
		tcpConn.SetNoDelay(true)
		client := &EasyConnection{conn: tcpConn, writeChan: make(chan []byte, EASYCALL_WRITE_QUEUE_SIZE), handler: handler, mutex: &sync.Mutex{}, activeTime: time.Now()}
		go client.Read()
		go client.Write()
	}
}
