package easycall

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/starjiang/elog"
)

type NodeInfo struct {
	name   string
	port   int
	weight int
}

type ServiceRegister struct {
	hostMap  map[string]*NodeInfo
	cli      *clientv3.Client
	leaseId  clientv3.LeaseID
	timeout  time.Duration
	register bool
}

func NewServiceRegister(endpoints []string, timeout time.Duration) (*ServiceRegister, error) {

	serviceRegister := &ServiceRegister{cli: nil, hostMap: make(map[string]*NodeInfo, 0), register: false}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: timeout,
	})

	if err != nil {
		return nil, err
	}

	serviceRegister.cli = cli
	serviceRegister.timeout = timeout

	return serviceRegister, nil

}

func (sr *ServiceRegister) Register(name string, port int, weight int) error {

	sr.Unregister(name, port)

	node := make(map[string]interface{}, 0)

	localIp := GetLocalIp()

	node["ip"] = localIp
	node["port"] = port
	node["weight"] = weight
	node["startTime"] = GetTimeNow()
	nodeInfo := &NodeInfo{name, port, weight}

	nodeData, err := json.Marshal(node)

	if err != nil {
		return err
	}

	nodeKey := EASYCALL_ETCD_SERVICE_PATH + "/" + name + "/nodes/" + localIp + ":" + strconv.Itoa(port)

	lease := clientv3.NewLease(sr.cli)
	ctx, cancel := context.WithTimeout(context.Background(), sr.timeout)
	lresp, err := lease.Grant(ctx, 10)
	cancel()
	if err != nil {
		return err
	}

	ctx, cancel = context.WithTimeout(context.Background(), sr.timeout)
	_, err = sr.cli.Put(ctx, nodeKey, string(nodeData), clientv3.WithLease(lresp.ID))
	cancel()
	if err != nil {
		return err
	}

	sr.leaseId = lresp.ID

	if !sr.register {
		go func() {
			for _ = range time.NewTicker(time.Second * time.Duration(ETCD_HEARTBEAT_INTEVAL)).C {
				ctx, cancel := context.WithTimeout(context.Background(), sr.timeout)
				_, err := lease.KeepAliveOnce(ctx, sr.leaseId)
				cancel()
				if err != nil {
					sr.Register(name, port, weight)
					elog.Error("send keepalive fail,", err)
					continue
				}
				elog.Info(name, "send etcd keepalive success")
			}
		}()
	}
	sr.hostMap[name] = nodeInfo
	sr.register = true
	return nil
}

func (sr *ServiceRegister) Unregister(name string, port int) error {

	localIp := GetLocalIp()

	nodeKey := EASYCALL_ETCD_SERVICE_PATH + "/" + name + "/nodes/" + localIp + ":" + strconv.Itoa(port)

	ctx, cancel := context.WithTimeout(context.Background(), sr.timeout)
	_, err := sr.cli.Delete(ctx, nodeKey)
	cancel()
	if err != nil {
		return err
	}

	delete(sr.hostMap, name)

	return nil
}
