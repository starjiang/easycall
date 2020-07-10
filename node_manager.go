package easycall

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/starjiang/elog"
)

const (
	ZK_NOT_EXSIT_NODE_CACHE_TIME = 5000
)

type NodeManager struct {
	exsit       int64
	mutex       *sync.Mutex
	cli         *clientv3.Client
	serviceName string
	nodeList    []*Node
	timeout     time.Duration
}

type Node struct {
	Ip     string `json:"ip"`
	Port   int    `json:"port"`
	Weight int    `json:"weight"`
	Active int32
}

func NewNodeManager(endpoints []string, serviceName string, timeout time.Duration) (*NodeManager, error) {
	nodeManager := &NodeManager{}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: timeout,
	})
	if err != nil {
		return nil, err
	}

	nodeManager.cli = cli
	nodeManager.serviceName = serviceName
	nodeManager.mutex = &sync.Mutex{}
	nodeManager.timeout = timeout
	return nodeManager, nil
}

func (nm *NodeManager) getNodes() ([]*Node, error) {

	path := EASYCALL_ETCD_SERVICE_PATH + "/" + nm.serviceName + "/nodes"

	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	if nm.nodeList == nil {
		timeNow := GetTimeNow()
		if nm.exsit+ZK_NOT_EXSIT_NODE_CACHE_TIME > timeNow {
			return nil, errors.New("zk node not exsit")
		}
		err := nm.loadServiceNode(path)
		if nm.nodeList == nil {
			nm.exsit = GetTimeNow()
			return nil, err
		}
		return nm.nodeList, nil
	}
	if len(nm.nodeList) == 0 {
		return nil, errors.New("service " + nm.serviceName + " not found")
	}
	return nm.nodeList, nil
}

func (nm *NodeManager) loadServiceNode(path string) error {

	ctx, cancel := context.WithTimeout(context.Background(), nm.timeout)
	resp, err := nm.cli.Get(ctx, path, clientv3.WithPrefix())
	cancel()

	if err != nil {
		return err
	}

	//children changed reload
	go func() {
		//watch
		rch := nm.cli.Watch(context.Background(), path, clientv3.WithPrefix())
		for _ = range rch {
			nm.loadServiceNode(path)
		}
		elog.Info("node watch failed")
	}()

	nodeList := make([]*Node, 0)

	for _, ev := range resp.Kvs {

		node := &Node{}
		err = json.Unmarshal(ev.Value, node)
		if err != nil {
			elog.Error("decode child data fail:", err)
			continue
		}
		nodeList = append(nodeList, node)
	}
	nm.nodeList = nodeList

	return nil
}
