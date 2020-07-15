package easycall

import (
	"errors"
	"hash/crc32"
	"math/rand"
	"time"
)

const (
	LB_ACTIVE        = 1
	LB_RANDOM        = 2
	LB_HASH          = 3
	LB_ROUND_ROBIN   = 4
	LB_RANDOM_WEIGHT = 5
)

type LoadBalancer struct {
	nodeList []*Node
	seq      int64
}

func NewLoadBalancer() *LoadBalancer {
	lb := &LoadBalancer{}
	return lb
}

func (lb *LoadBalancer) SetNodes(nodeList []*Node) {
	lb.nodeList = nodeList
}

func (lb *LoadBalancer) GetNode(loadBalanceType int, routeKey string) (*Node, error) {

	var node *Node = nil
	if len(lb.nodeList) == 0 {
		return nil, errors.New("service not found")
	}

	if loadBalanceType == LB_ACTIVE {
		node = lb.getNodeByLoadBalanceActive()
	} else if loadBalanceType == LB_RANDOM {
		node = lb.getNodeByLoadBalanceRandom()
	} else if loadBalanceType == LB_RANDOM_WEIGHT {
		node = lb.getNodeByLoadBalanceRandomWeight()
	} else if loadBalanceType == LB_ROUND_ROBIN {
		node = lb.getNodeByLoadBalanceRoundRobin()
	} else if loadBalanceType == LB_HASH {
		node = lb.getNodeByLoadBalanceHash(routeKey)
	} else {
		return nil, errors.New("invalid loadBalanceType")
	}
	return node, nil
}

func (lb *LoadBalancer) getNodeByLoadBalanceActive() *Node {

	len := len(lb.nodeList)
	index := 0
	active := lb.nodeList[0].Active
	for i := 1; i < len; i++ {
		if active >= lb.nodeList[i].Active {
			active = lb.nodeList[i].Active
			index = i
		}
	}

	if index == (len - 1) {
		return lb.getNodeByLoadBalanceRoundRobin()
	}

	return lb.nodeList[index]
}

func (lb *LoadBalancer) getNodeByLoadBalanceRandom() *Node {

	len := len(lb.nodeList)
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len)
	return lb.nodeList[index]
}

func (lb *LoadBalancer) hashKey(key string) uint32 {
	if len(key) < 64 {
		var scratch [64]byte
		copy(scratch[:], key)
		return crc32.ChecksumIEEE(scratch[:len(key)])
	}
	return crc32.ChecksumIEEE([]byte(key))
}

func (lb *LoadBalancer) getNodeByLoadBalanceHash(routeKey string) *Node {

	len := len(lb.nodeList)
	index := lb.hashKey(routeKey) % uint32(len)
	return lb.nodeList[index]
}

func (lb *LoadBalancer) getNodeByLoadBalanceRandomWeight() *Node {

	total := 0
	for i := 0; i < len(lb.nodeList); i++ {
		node := lb.nodeList[i]
		total += node.Weight
	}
	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(total)
	for i := 0; i < len(lb.nodeList); i++ {
		node := lb.nodeList[i]
		random -= node.Weight
		if random <= 0 {
			return node
		}
	}
	return lb.nodeList[0]
}

func (lb *LoadBalancer) getNodeByLoadBalanceRoundRobin() *Node {
	lb.seq++
	len := len(lb.nodeList)
	index := int(lb.seq % int64(len))
	return lb.nodeList[index]
}
