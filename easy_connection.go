package easycall

import (
	"encoding/binary"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/starjiang/elog"
)

type EasyConnection struct {
	conn       *net.TCPConn
	isClose    bool
	writeChan  chan []byte
	handler    PkgHandler
	activeTime time.Time
	mutex      *sync.Mutex
}

func (ec *EasyConnection) Close() error {
	if ec.isClose == false {
		ec.isClose = true
		return ec.conn.Close()
	}
	return nil
}

func (ec *EasyConnection) GetTcpConn() *net.TCPConn {
	return ec.conn
}

func (ec *EasyConnection) Send(pkgData []byte) {

	if ec.isClose == true {
		elog.Info("tcp connection is closed,can't send pkg")
		return
	}
	if ec.writeChan != nil {
		ec.writeChan <- pkgData
	}
}

func (ec *EasyConnection) IsClose() bool {
	return ec.isClose
}

func (ec *EasyConnection) Write() {

	defer PanicHandler()

	for {
		select {
		case pkgData, ok := <-ec.writeChan:
			if !ok {
				elog.Error("get respPkg from write channel fail,maybe channel closed")
				return
			} else {
				if pkgData == nil {
					elog.Info("exit write process")
					return
				}
				_, err := ec.conn.Write(pkgData)
				if err != nil {
					if err == io.EOF {
						elog.Info(ec.conn.LocalAddr().String(), ec.conn.RemoteAddr().String(), "conn closed")
					} else {
						elog.Error(ec.conn.LocalAddr().String(), ec.conn.RemoteAddr().String(), "conn write exception:", err)
					}
					return
				}
			}
		}
	}
}

func (ec *EasyConnection) Read() {
	defer func() {
		ec.Close()
		if ec.writeChan != nil {
			ec.writeChan <- nil
			close(ec.writeChan)
		}
	}()

	defer PanicHandler()

	for {
		var prefetch = make([]byte, 10)
		_, err := ec.conn.Read(prefetch)
		if err != nil {
			if err == io.EOF {
				elog.Info(ec.conn.LocalAddr().String(), ec.conn.RemoteAddr().String(), "conn closed")
			} else {
				elog.Error(ec.conn.LocalAddr().String(), ec.conn.RemoteAddr().String(), "conn read exception:", err)
			}
			return
		}

		stx := prefetch[0]
		format := prefetch[1]
		headLen := binary.BigEndian.Uint32(prefetch[2:6])
		bodyLen := binary.BigEndian.Uint32(prefetch[6:10])

		if stx != STX {
			elog.Error("invalid pkg stx", stx)
			return
		}
		if format != FORMAT_JSON && format != FORMAT_MSGPACK {
			elog.Error("invalid pkg format")
			return
		}
		if headLen > HEAD_MAX_LEN {
			elog.Error("invalid pkg headlen", headLen)
			return
		}
		if headLen > BODY_MAX_LEN {
			elog.Error("invalid pkg bodylen", bodyLen)
			return
		}
		pkgLen := 11 + headLen + bodyLen
		var pkgData = make([]byte, pkgLen)
		copy(pkgData, prefetch)
		_, err = ec.conn.Read(pkgData[10:])
		if err != nil {
			if err == io.EOF {
				elog.Info(ec.conn.LocalAddr().String(), ec.conn.RemoteAddr().String(), "conn closed")
			} else {
				elog.Error(ec.conn.LocalAddr().String(), ec.conn.RemoteAddr().String(), "conn read exception:", err)
			}
			return
		}
		if pkgData[pkgLen-1] != ETX {
			elog.Error("invalid pkg etx", pkgData[pkgLen-1])
			return
		}
		go ec.handler.Dispatch(pkgData, ec)
	}
}

func (ec *EasyConnection) GetActiveTime() time.Time {
	return ec.activeTime
}

func (ec *EasyConnection) Connect(ip string, port int) error {
	conn, err := net.DialTimeout("tcp", ip+":"+strconv.Itoa(port), time.Second*EASY_CONNECT_TIMEOUT)
	if err != nil {
		return err
	}
	ec.conn = conn.(*net.TCPConn)
	ec.conn.SetKeepAlive(true)
	ec.conn.SetKeepAlivePeriod(time.Minute * TCP_KEEPALIVE_PERIOD)
	ec.conn.SetNoDelay(true)
	ec.activeTime = time.Now()
	ec.writeChan = make(chan []byte, EASYCALL_WRITE_QUEUE_SIZE)
	go ec.Read()
	go ec.Write()
	ec.isClose = false
	return nil
}
