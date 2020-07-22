package easycall

import (
	"sync"
	"sync/atomic"
	"time"
)

const APM_COUNT_INTERVAL = 60

type ApmMonitorHandler interface {
	OnData(data map[string]*ApmMonitorStatus)
}

func NewApmMonitor(handler ApmMonitorHandler) *ApmMonitor {
	apmMonitor := &ApmMonitor{}
	apmMonitor.handler = handler
	apmMonitor.interval = APM_COUNT_INTERVAL * time.Second
	apmMonitor.mutex = &sync.Mutex{}
	apmMonitor.statuses = make(map[string]*ApmMonitorStatus)
	go apmMonitor.reportAndReset()
	return apmMonitor
}

type ApmMonitorStatus struct {
	Total uint64
	Error uint64
	Time  uint64
}

type ApmMonitor struct {
	handler  ApmMonitorHandler
	interval time.Duration
	statuses map[string]*ApmMonitorStatus
	mutex    *sync.Mutex
}

func (am *ApmMonitor) Middleware(req *Request, resp *Response, client *EasyConnection, next *MiddlewareInfo) {
	am.mutex.Lock()
	status := am.statuses[req.GetHead().GetMethod()]
	if status == nil {
		status = &ApmMonitorStatus{}
		am.statuses[req.GetHead().GetMethod()] = status
	}
	am.mutex.Unlock()

	atomic.AddUint64(&status.Total, 1)

	start := time.Now()
	next.Middleware(req, resp, client, next.Next)
	end := time.Now()

	spendTime := end.Sub(start)
	atomic.AddUint64(&status.Time, uint64(spendTime.Milliseconds()))

	head := resp.GetHead()
	if head != nil {
		if head.GetRet() != 0 {
			atomic.AddUint64(&status.Error, 1)
		}
	}
}

func (am *ApmMonitor) reportAndReset() {

	for range time.NewTicker(am.interval).C {

		am.mutex.Lock()
		statuses := am.statuses
		am.statuses = make(map[string]*ApmMonitorStatus)
		am.mutex.Unlock()

		for _, v := range statuses {
			if v.Total > 0 {
				v.Time = v.Time / v.Total
			}
		}
		am.handler.OnData(statuses)
	}
}
