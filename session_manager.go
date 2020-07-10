package easycall

import (
	"sync"
	"sync/atomic"
	"time"
)

type EasySession struct {
	seqOrgi  uint64
	seq      uint64
	respChan chan *EasyPackage
	timer    *time.Timer
	mutex    *sync.Mutex
	node     *Node
}

type EasySessionManager struct {
	sessionMap map[uint64]*EasySession
	mutex      *sync.RWMutex
	seq        uint64
}

func (esm *EasySessionManager) AddSession(sessionId uint64, session *EasySession) {
	esm.mutex.Lock()
	esm.sessionMap[sessionId] = session
	esm.mutex.Unlock()
}

func (esm *EasySessionManager) RemoveSession(sessionId uint64) {
	esm.mutex.Lock()
	delete(esm.sessionMap, sessionId)
	esm.mutex.Unlock()
}

func (esm *EasySessionManager) GetSession(sessionId uint64) *EasySession {
	esm.mutex.RLock()
	easySession := esm.sessionMap[sessionId]
	esm.mutex.RUnlock()
	return easySession
}

func (esm *EasySessionManager) InitSession(timeout time.Duration, node *Node) *EasySession {

	seq := atomic.AddUint64(&esm.seq, 1)
	atomic.AddInt32(&node.Active, 1)

	session := &EasySession{0, seq, make(chan *EasyPackage), time.NewTimer(timeout), &sync.Mutex{}, node}

	esm.mutex.Lock()
	esm.sessionMap[seq] = session
	esm.mutex.Unlock()
	return session
}

func (esm *EasySessionManager) DestorySessionAndRespPkg(session *EasySession, respPkg *EasyPackage) {

	if session == nil {
		return
	}

	esm.mutex.Lock()
	defer esm.mutex.Unlock()

	_, ok := esm.sessionMap[session.seq]

	if !ok {
		return
	}

	atomic.AddInt32(&session.node.Active, -1)

	if session.respChan != nil {
		session.respChan <- respPkg
		close(session.respChan)
		session.respChan = nil
	}
	if session.timer != nil {
		session.timer.Stop()
	}

	delete(esm.sessionMap, session.seq)

}
