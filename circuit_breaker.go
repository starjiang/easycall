package easycall

import (
	"math/rand"
	"sync"
	"sync/atomic"

	"github.com/starjiang/elog"
)

const (
	CB_FAIL_RATE    = 0.5
	CB_LIMIT_RATE   = 0.2
	CB_COUNT_BASE   = 10
	CB_FAIL_TIME    = 30000
	CB_LIMIT_TIME   = 30000
	CB_RESET_TIME   = 60000
	CB_STATUS_OPEN  = 1
	CB_STATUS_CLOSE = 0
	CB_STATUS_LIMIT = 2
)

type runFunc func() error
type failFunc func() error

type CbInfo struct {
	invokeCount            int64
	failCount              int64
	lastCircuitBreakerTime int64
	lastLimitPassTime      int64
	lastResetTime          int64
	status                 int
	failRate               float32
	limitRate              float32
	countBase              int
	failTime               int64
	limitTime              int64
	resetTime              int64
}

var cbInfos = make(map[string]*CbInfo)
var cbMutexes = make(map[string]*sync.Mutex)

var cbMutex sync.Mutex

func CbConfigure(cbName string, failRate float32, limitRate float32, countBase int, failTime int64, limitTime int64, resetTime int64) {

	timeNow := GetTimeNow()

	var lock *sync.Mutex
	cbMutex.Lock()
	lock = cbMutexes[cbName]
	if lock == nil {
		lock = &sync.Mutex{}
		cbMutexes[cbName] = lock
	}

	info := cbInfos[cbName]
	if info == nil {
		info = &CbInfo{1, 0, 0, 0, timeNow, CB_STATUS_CLOSE, failRate, limitRate, countBase, failTime, limitTime, resetTime}
		cbInfos[cbName] = info
	}

	cbMutex.Unlock()

	lock.Lock()
	info.failRate = failRate
	info.limitRate = limitRate
	info.countBase = countBase
	info.failTime = failTime
	info.limitTime = limitTime
	info.resetTime = resetTime
	lock.Unlock()

}

func CbCall(cbName string, run runFunc, fail failFunc) error {

	timeNow := GetTimeNow()

	var lock *sync.Mutex
	cbMutex.Lock()
	lock = cbMutexes[cbName]
	if lock == nil {
		lock = &sync.Mutex{}
		cbMutexes[cbName] = lock
	}

	info := cbInfos[cbName]
	if info == nil {
		info = &CbInfo{1, 0, 0, 0, timeNow, CB_STATUS_CLOSE, CB_FAIL_RATE, CB_LIMIT_RATE, CB_COUNT_BASE, CB_FAIL_TIME, CB_LIMIT_TIME, CB_RESET_TIME}
		cbInfos[cbName] = info
	}

	cbMutex.Unlock()

	lock.Lock()

	//熔断过期后，把熔断器状态设置为半开状
	if info.status == CB_STATUS_OPEN && info.lastCircuitBreakerTime+info.failTime < timeNow {
		info.status = CB_STATUS_LIMIT
		info.failCount = 0
		info.invokeCount = 1
		info.lastLimitPassTime = timeNow
		info.lastCircuitBreakerTime = 0
		elog.Errorf("CircuitBreaker %s set status limit", cbName)
	}

	ptFlag := false
	if info.status == CB_STATUS_OPEN {
		ptFlag = false
	} else if info.status == CB_STATUS_LIMIT {
		//半开状态下，随机计算允许通过的请求
		rand := rand.Float32()
		if rand < info.limitRate {
			ptFlag = true
		}
	} else {
		ptFlag = true
	}

	if !ptFlag {
		if fail != nil {
			lock.Unlock()
			return fail()
		}
		lock.Unlock()
		return nil
	}

	//计算熔断阀值，超过阀值，熔断
	f := float32(info.failCount) / float32(info.invokeCount)

	if f > info.failRate && info.invokeCount > int64(info.countBase) {
		info.status = CB_STATUS_OPEN
		info.lastCircuitBreakerTime = timeNow
		elog.Errorf("CircuitBreaker %s set status open", cbName)
		if fail != nil {
			lock.Unlock()
			return fail()
		}
		lock.Unlock()
		return nil
	}

	//半开状态下，请求没超过阀值，关闭熔断器
	if info.status == CB_STATUS_LIMIT && info.lastLimitPassTime+info.limitTime < timeNow {
		info.status = CB_STATUS_CLOSE
		info.invokeCount = 1
		info.failCount = 0
		info.lastLimitPassTime = 0
		elog.Errorf("CircuitBreaker %s set status close", cbName)
	}

	//重置统计
	if info.status == CB_STATUS_CLOSE && info.lastResetTime+info.resetTime < timeNow {
		info.lastResetTime = timeNow
		info.failCount = 0
		info.invokeCount = 1
		elog.Infof("CircuitBreaker %s reset", cbName)
	}

	lock.Unlock()

	atomic.AddInt64(&info.invokeCount, 1)

	err := run()

	if err != nil {
		atomic.AddInt64(&info.failCount, 1)
	}
	return err
}
