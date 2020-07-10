package easycall

import (
	"errors"
	"io"
	"sync/atomic"
	"time"
)

type Poolable interface {
	io.Closer
	GetActiveTime() time.Time
	IsClose() bool
}

type factory func() (Poolable, error)

type Pool interface {
	Acquire() (Poolable, error) // 获取资源
	Release(Poolable) error     // 释放资源
	Close(Poolable) error       // 关闭资源
	Shutdown() error            // 关闭池
}

type GenericPool struct {
	poolChan    chan Poolable
	maxSize     int32 // 池中最大资源数
	curSize     int32 // 当前池中资源数
	minSize     int32 // 池中最少资源数
	shutdown    bool  // 池是否已关闭
	lifetime    time.Duration
	connFactory factory // 创建连接的方法
}

func NewGenericPool(minSize int32, maxSize int32, lifetime time.Duration, connFactory factory) *GenericPool {
	if maxSize <= 0 || minSize > maxSize {
		maxSize = minSize
	}
	pool := &GenericPool{
		maxSize:     maxSize,
		minSize:     minSize,
		lifetime:    lifetime,
		curSize:     0,
		connFactory: connFactory,
		poolChan:    make(chan Poolable, maxSize*2),
	}

	for i := 0; i < int(minSize); i++ {
		conn, err := connFactory()
		if err != nil {
			continue
		}
		atomic.AddInt32(&pool.curSize, 1)
		pool.poolChan <- conn
	}
	return pool
}

func (pool *GenericPool) Acquire() (Poolable, error) {
	if pool.shutdown {
		return nil, errors.New("pool have been shutdown")
	}

	for {
		conn, err := pool.getOrCreate()

		if err != nil {
			return nil, err
		}

		if conn.IsClose() {
			atomic.AddInt32(&pool.curSize, -1)
			continue
		}
		// 如果设置了超时且当前连接的活跃时间+超时时间早于现在，则当前连接已过期
		if pool.lifetime > 0 && conn.GetActiveTime().Add(time.Duration(pool.lifetime)).Before(time.Now()) {
			pool.Close(conn)
			continue
		}
		return conn, nil
	}
}

func (pool *GenericPool) getOrCreate() (Poolable, error) {

	if pool.shutdown {
		return nil, errors.New("pool have been shutdown")
	}

	select {
	case conn := <-pool.poolChan:
		return conn, nil
	default:
	}

	if pool.curSize >= pool.maxSize {
		timeout := make(chan bool, 1)
		go func() {
			time.Sleep(time.Second * POOL_MAX_WAIT_TIME)
			timeout <- true
		}()

		select {
		case conn := <-pool.poolChan:
			return conn, nil
		case <-timeout:
		}
		return nil, errors.New("no connection available")
	}
	// 新建连接
	conn, err := pool.connFactory()
	if err != nil {
		return nil, err
	}
	atomic.AddInt32(&pool.curSize, 1)
	return conn, nil
}

// 释放单个资源到连接池
func (p *GenericPool) Release(conn Poolable) error {
	if p.shutdown {
		return errors.New("pool have benn shutdown")
	}
	p.poolChan <- conn
	return nil
}

// 关闭单个资源
func (pool *GenericPool) Close(conn Poolable) error {
	conn.Close()
	atomic.AddInt32(&pool.curSize, -1)
	return nil
}

func (pool *GenericPool) IsShutDown() bool {
	return pool.shutdown
}

func (pool *GenericPool) CloseAll() {
	if pool.shutdown {
		return
	}

	for {
		select {
		case conn := <-pool.poolChan:
			conn.Close()
			atomic.AddInt32(&pool.curSize, -1)
		default:
			return
		}
	}
}

// 关闭连接池，释放所有资源
func (pool *GenericPool) Shutdown() error {
	if pool.shutdown {
		return errors.New("pool have been shutdown")
	}
	pool.shutdown = true

	close(pool.poolChan)

	for conn := range pool.poolChan {
		conn.Close()
		atomic.AddInt32(&pool.curSize, -1)
	}
	return nil
}
