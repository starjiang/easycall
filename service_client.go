package easycall

import (
	"strconv"
	"sync"
	"time"

	"github.com/starjiang/elog"
)

const (
	POOL_MIN_SIZE        = 10
	POOL_ACTIVE_TIME     = 1800
	POOL_MAX_WAIT_TIME   = 5
	EASY_CONNECT_TIMEOUT = 3
)

type ServiceClient struct {
	sessionMgr      *EasySessionManager
	nodeMgr         *NodeManager
	poolMap         map[string]*GenericPool
	mutex           *sync.Mutex
	loadBalanceType int
	poolSize        int
	seq             uint64
	serviceName     string
	lb              *LoadBalancer
}

//create a new service request client

//endpoints etcd endpoints list
//serviceName microservice name
//poolsize connection pool size
//loadBalanceType for 5 kinds of loadbalance
func NewServiceClient(endpoints []string, serviceName string, poolSize int, loadBalanceType int) *ServiceClient {

	ServiceClient := &ServiceClient{}
	nodeMgr, err := NewNodeManager(endpoints, serviceName, ETCD_CONNECT_TIMEOUT*time.Second)
	if err != nil {
		elog.Error("new nodemgr fail:", err)
	}
	ServiceClient.nodeMgr = nodeMgr
	ServiceClient.sessionMgr = &EasySessionManager{sessionMap: make(map[uint64]*EasySession, 0), mutex: &sync.RWMutex{}}
	ServiceClient.poolMap = make(map[string]*GenericPool, 0)
	ServiceClient.mutex = &sync.Mutex{}
	ServiceClient.loadBalanceType = loadBalanceType
	ServiceClient.serviceName = serviceName
	ServiceClient.lb = NewLoadBalancer()
	return ServiceClient
}

//request with head

//format serialize format type json/msgpack
//head request head
//body request body
//timeout request timeout
func (ec *ServiceClient) RequestWithHead(format byte, head *EasyHead, body interface{}, timeout time.Duration) (*EasyPackage, error) {

	ch, err := ec.RequestAsyncWithHead(format, head, body, timeout)

	if err != nil {
		return nil, err
	}
	respPkg := <-ch
	if respPkg == nil {
		return nil, NewSystemError(ERROR_TIME_OUT, "request time out")
	}
	return respPkg, nil
}

func (ec *ServiceClient) Request(method string, reqBody interface{}, respBody interface{}, timeout time.Duration) error {

	ch, err := ec.RequestAsyncWithHead(FORMAT_MSGPACK, NewEasyHead().SetService(ec.serviceName).SetMethod(method), reqBody, timeout)

	if err != nil {
		return err
	}
	respPkg := <-ch
	if respPkg == nil {
		return NewSystemError(ERROR_TIME_OUT, "request time out")
	}
	if respPkg.GetHead().GetRet() < ERROR_MAX_SYSTEM_CODE {
		return NewSystemError(respPkg.GetHead().GetRet(), respPkg.GetHead().GetMsg())
	} else {
		return NewLogicError(respPkg.GetHead().GetRet(), respPkg.GetHead().GetMsg())
	}
	return respPkg.DecodeBody(respBody)

}

func (ec *ServiceClient) RequestAsync(method string, body interface{}, timeout time.Duration) (chan *EasyPackage, error) {
	return ec.RequestAsyncWithHead(FORMAT_MSGPACK, NewEasyHead().SetService(ec.serviceName).SetMethod(method), body, timeout)
}

//request with head through by Async

//format serialize format type json/msgpack
//head request head
//body request body
//timeout request timeout
func (ec *ServiceClient) RequestAsyncWithHead(format byte, head *EasyHead, body interface{}, timeout time.Duration) (chan *EasyPackage, error) {

	if head.GetService() != ec.serviceName {
		return nil, NewSystemError(ERROR_INTERNAL_ERROR, "service name is different from init")
	}

	if ec.nodeMgr == nil {
		return nil, NewSystemError(ERROR_INTERNAL_ERROR, "nodemgr is nil,maybe etcd connect fail")
	}
	lbType := ec.loadBalanceType

	if head.GetRouteKey() != "" {
		lbType = LB_HASH
	}

	nodeList, err := ec.nodeMgr.getNodes()

	if err != nil {
		return nil, NewSystemError(ERROR_SERVICE_NOT_FOUND, err.Error())
	}

	ec.lb.SetNodes(nodeList)
	node, err := ec.lb.GetNode(lbType, head.GetRouteKey())
	if err != nil {
		return nil, NewSystemError(ERROR_INTERNAL_ERROR, err.Error())
	}
	key := node.Ip + ":" + strconv.Itoa(node.Port)

	ec.mutex.Lock()
	var pool *GenericPool
	if ec.poolMap[key] == nil {
		pool = NewGenericPool(int32(POOL_MIN_SIZE), int32(ec.poolSize), time.Second*POOL_ACTIVE_TIME, func() (Poolable, error) {
			clientHandler := NewClientHandler(ec)
			conn := &EasyConnection{conn: nil, isClose: true, writeChan: nil, handler: clientHandler, activeTime: time.Now(), mutex: &sync.Mutex{}}
			err := conn.Connect(node.Ip, node.Port)
			if err != nil {
				return nil, err
			}
			return conn, err
		})
		ec.poolMap[key] = pool
	} else {
		pool = ec.poolMap[key]
	}
	ec.mutex.Unlock()

	conn, err := pool.Acquire()
	if err != nil {
		return nil, NewSystemError(ERROR_INTERNAL_ERROR, err.Error())
	}
	pool.Release(conn)

	easyConn := conn.(*EasyConnection)

	session := ec.sessionMgr.InitSession(timeout, node)
	head.SetSeq(session.seq)

	bodyData, ok := body.([]byte)
	if ok {

		pkgData, err := NewPackageWithBodyData(format, head, bodyData).EncodeWithBodyData()
		if err != nil {
			ec.sessionMgr.DestorySessionAndRespPkg(session, nil)
			return nil, NewSystemError(ERROR_INTERNAL_ERROR, err.Error())
		}
		easyConn.Send(pkgData)

	} else {
		pkgData, err := NewPackageWithBody(format, head, body).EncodeWithBody()

		if err != nil {
			ec.sessionMgr.DestorySessionAndRespPkg(session, nil)
			return nil, NewSystemError(ERROR_INTERNAL_ERROR, err.Error())
		}
		easyConn.Send(pkgData)
	}

	go func() {
		<-session.timer.C
		ec.sessionMgr.DestorySessionAndRespPkg(session, nil)
	}()

	return session.respChan, nil
}

func (ec *ServiceClient) Process(respPkg *EasyPackage) {

	if respPkg.GetHead().GetSeq() == 0 {
		elog.Errorf("pkg service=%s,method=%s head seq not setup", respPkg.GetHead().GetService(), respPkg.GetHead().GetMethod())
		return
	}

	session := ec.sessionMgr.GetSession(respPkg.GetHead().GetSeq())

	if session == nil {
		elog.Errorf("pkg service=%s,method=%s session not found", respPkg.GetHead().GetService(), respPkg.GetHead().GetMethod())
		return
	}
	ec.sessionMgr.DestorySessionAndRespPkg(session, respPkg)

}
