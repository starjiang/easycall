package easycall

import (
	"sync"
	"time"
)

type EasyClient struct {
	clients         map[string]*ServiceClient
	mutex           *sync.Mutex
	endpoints       []string
	poolSize        int
	loadbalanceType int
}

func NewEasyClient(endpoints []string, poolSize int, loadbalanceType int) *EasyClient {
	return &EasyClient{endpoints: endpoints, mutex: &sync.Mutex{}, clients: make(map[string]*ServiceClient, 0), poolSize: poolSize, loadbalanceType: loadbalanceType}
}

func (ec *EasyClient) Request(serviceName string, method string, reqBody interface{}, respBody interface{}, timeout time.Duration) error {

	ch, err := ec.RequestAsyncWithHead(FORMAT_MSGPACK, NewEasyHead().SetService(serviceName).SetMethod(method), reqBody, timeout)

	if err != nil {
		return err
	}
	respPkg := <-ch
	if respPkg == nil {
		return NewSystemError(ERROR_TIME_OUT, "request time out")
	}
	if respPkg.GetHead().GetRet() != 0 {
		return NewLogicError(respPkg.GetHead().GetRet(), respPkg.GetHead().GetMsg())
	}
	return respPkg.DecodeBody(respBody)
}

func (ec *EasyClient) RequestAsync(serviceName string, method string, body interface{}, timeout time.Duration) (chan *EasyPackage, error) {
	return ec.RequestAsyncWithHead(FORMAT_MSGPACK, NewEasyHead().SetService(serviceName).SetMethod(method), body, timeout)
}

func (ec *EasyClient) RequestWithHead(format byte, head *EasyHead, body interface{}, timeout time.Duration) (*EasyPackage, error) {

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

func (ec *EasyClient) RequestAsyncWithHead(format byte, head *EasyHead, reqBody interface{}, timeout time.Duration) (chan *EasyPackage, error) {

	ec.mutex.Lock()
	client := ec.clients[head.GetService()]
	if client == nil {
		client = NewServiceClient(ec.endpoints, head.GetService(), ec.poolSize, ec.loadbalanceType)
		ec.clients[head.GetService()] = client
	}
	ec.mutex.Unlock()

	return client.RequestAsyncWithHead(format, head, reqBody, timeout)
}
