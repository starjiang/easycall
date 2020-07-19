package easycall

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/starjiang/elog"
)

//ServiceInfo for ServiceContext
type ServiceInfo struct {
	name    string
	port    int
	weight  int
	service interface{}
}

type MiddlewareFunc func(req *Request, resp *Response, client *EasyConnection, next *MiddlewareInfo)

type MiddlewareInfo struct {
	Middleware MiddlewareFunc
	Next       *MiddlewareInfo
}

//ServiceContext for Easycall
type ServiceContext struct {
	serviceList map[string]*ServiceInfo
	endpoints   []string
	middlewares map[string][]*MiddlewareInfo
}

//endpoints etcd endpoints list
func NewServiceContext(endpoints []string) *ServiceContext {

	return &ServiceContext{make(map[string]*ServiceInfo, 0), endpoints, make(map[string][]*MiddlewareInfo, 0)}
}

//name microservice name
//port microservice port
//service microservice implement
//weight microservice weight for loadbalance
func (svc *ServiceContext) CreateService(name string, port int, service interface{}, weight int) error {
	info := &ServiceInfo{name, port, weight, service}
	svc.serviceList[name] = info
	return nil
}

//name microservice name
//middleware fucntion for middleware function chain
func (svc *ServiceContext) AddMiddleware(name string, middleware MiddlewareFunc) {
	list := svc.middlewares[name]
	if list == nil {
		list = make([]*MiddlewareInfo, 0)
	}
	minfo := &MiddlewareInfo{middleware, nil}

	mlen := len(list)
	if mlen > 0 {
		list[mlen-1].Next = minfo
	}

	list = append(list, minfo)
	svc.middlewares[name] = list
}

//register and start all microservices and wait
func (svc *ServiceContext) StartAndWait() error {

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				for _, info := range svc.serviceList {
					elog.Infof("unregister service %s,port=%d", info.name, info.port)
					register, err := NewServiceRegister(svc.endpoints, time.Second*ETCD_CONNECT_TIMEOUT)
					if err != nil {
						continue
					}
					register.Unregister(info.name, info.port)
				}
				os.Exit(0)
			default:
				elog.Info("other signal:", s)
			}
		}
	}()

	var wg sync.WaitGroup
	size := len(svc.serviceList)
	wg.Add(size)
	for _, info := range svc.serviceList {
		server := &Server{}
		go func(info *ServiceInfo, wg *sync.WaitGroup) {
			elog.Infof("service %s start at port %d", info.name, info.port)
			handler := NewServiceHandler(info.service, svc.middlewares[info.name])
			err := server.CreateServer(info.port, handler)
			if err != nil {
				elog.Error("start service fail:", err, info.name, info.port, info.weight)
				wg.Done()
				return
			}
			register, err := NewServiceRegister(svc.endpoints, time.Second*ETCD_CONNECT_TIMEOUT)
			if err != nil {
				wg.Done()
				elog.Error("init register fail:", err, info.name, info.port, info.weight)
				return
			}
			err = register.Register(info.name, info.port, info.weight)
			if err != nil {
				wg.Done()
				elog.Error("register fail:", err, info.name, info.port, info.weight)
				return
			}

		}(info, &wg)
	}

	wg.Wait()

	return nil
}
