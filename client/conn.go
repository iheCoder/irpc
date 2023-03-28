package client

import (
	"crypto/tls"
	"errors"
	"github.com/lucas-clemente/quic-go"
	"io"
	"learn/irpc/common"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type StreamConn interface {
	io.Reader
	io.Writer
	io.Closer
}

const (
	maxOpenStream = 50
	maxConn       = 100

	defaultKeepAlivePeriod = time.Second

	idle  = 0
	using = 1
)

var (
	MaxLastUseTime = time.Date(3000, 1, 1, 0, 0, 0, 0, time.Local)

	expireDuration = 10 * time.Second
)

var (
	ErrExceedConnMax   = errors.New("irpcClient conn: exceed conn max")
	ErrExceedStreamMax = errors.New("irpcClient conn: exceed stream max")
)

type AdapterConn struct {
	tlsConfig         *tls.Config
	dialAddr          string
	conns             []*ConnInfo
	streamSizePerConn int
	maxConnLen        int
	ticker            *time.Ticker
	mu                *sync.Mutex
	cfg               *quic.Config
}

func NewAdapterConn(tlsConfig *tls.Config, dialAddr string, ed, scanDuration time.Duration) *AdapterConn {
	ac := &AdapterConn{
		tlsConfig:         tlsConfig,
		dialAddr:          dialAddr,
		conns:             make([]*ConnInfo, 0),
		streamSizePerConn: maxOpenStream,
		ticker:            time.NewTicker(scanDuration),
		maxConnLen:        maxConn,
		mu:                &sync.Mutex{},
		cfg:               &quic.Config{KeepAlivePeriod: defaultKeepAlivePeriod},
	}

	if ed != time.Duration(0) {
		expireDuration = ed
	}

	go ac.cleanConn()

	return ac
}

// AcquireStream 获取stream
func (c *AdapterConn) AcquireStream() (StreamConn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 尝试获取stream
	ci, si, err := c.getConnStream()
	if err != nil {
		return nil, err
	}

	// 要确保streamCount也是和tryGetConn在同一原子中，不然还是open太多stream
	// 诡异，居然只创建了一个conn，这不应该啊。甚至连遍历都没有就too many open streams panic
	ci.rwMutex.Lock()
	ci.lastUseTime = MaxLastUseTime
	si.flag.Store(using)
	ci.rwMutex.Unlock()

	// construct StreamConn
	sc := &AdapterStreamConn{
		ci: ci,
		si: si,
	}

	return sc, nil
}

func (c *AdapterConn) getConnStream() (*ConnInfo, *StreamInfo, error) {
	// 若第一个conn不存在，则连接并获取open stream
	if len(c.conns) == 0 {
		if len(c.conns) == 0 {
			ci, err := c.createConn()
			if err != nil {
				return nil, nil, err
			}
			c.conns = append(c.conns, ci)

			si, err := ci.tryGetStream()
			if err != nil {
				return nil, nil, err
			}

			return ci, si, nil
		}
	}

	// 从头遍历连接，若找到可用连接就直接用吧
	// 很有可能很多线程到达这里第一个确实超出
	// 要不所有的都锁住怎么样，确实防止非常多的conn啦
	for _, ci := range c.conns {
		si, err := ci.tryGetStream()
		if err == ErrExceedStreamMax {
			continue
		}
		if err != nil {
			return nil, nil, err
		}

		return ci, si, nil
	}

	// 若超出最大限制，则返回err
	if len(c.conns) >= c.maxConnLen {
		return nil, nil, ErrExceedConnMax
	}

	// 若都没有可用的stream，那就创建新的吧
	ci, err := c.createConn()
	if err != nil {
		return nil, nil, err
	}

	c.conns = append(c.conns, ci)
	si, err := ci.tryGetStream()
	if err != nil {
		return nil, nil, err
	}

	return ci, si, nil
}

func (c *AdapterConn) createConn() (*ConnInfo, error) {
	conn, err := quic.DialAddr(c.dialAddr, c.tlsConfig, c.cfg)
	if err != nil {
		log.Printf("AdapterConn createConn: dial addr %s failed %s", c.dialAddr, err)
		return nil, err
	}

	return &ConnInfo{
		conn:           conn,
		lastUseTime:    MaxLastUseTime,
		rwMutex:        &sync.RWMutex{},
		streams:        make([]*StreamInfo, 0),
		maxStreamCount: c.streamSizePerConn,
	}, nil
}

func (c *AdapterConn) cleanConn() {
	for range c.ticker.C {
		c.mu.Lock()
		for i, connInfo := range c.conns {
			// 会不会存在connInfo仍然还在里面放stream，或者获取stream进行使用呢？在getStream通过AdapterConn锁锁住的时候是不存在的
			// 也不可能出现此时还能close啊，因为还能close就说明还有可用的stream
			if !connInfo.existsAvailableStream() && time.Now().After(connInfo.lastUseTime) {
				err := connInfo.conn.CloseWithError(common.TooLongToUsedErrCode, "conn hasn't been used for too long")
				if err != nil {
					c.mu.Unlock()
					log.Printf("AdapterConn cleanConn: close %s failed %s", c.dialAddr, err)
					return
				}
				c.conns = append(c.conns[:i], c.conns[i+1:]...)
			}
		}
		c.mu.Unlock()
	}
}

type ConnInfo struct {
	conn        quic.Connection
	streams     []*StreamInfo
	lastUseTime time.Time
	// 记录stream个数似乎是极为容易冲突的，那么为什么用互斥锁呢
	rwMutex        *sync.RWMutex
	maxStreamCount int
}

func (c *ConnInfo) tryGetStream() (*StreamInfo, error) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()
	// 寻找是否存在空闲的stream
	for _, si := range c.streams {
		if si.flag.Load().(int) == idle {
			return si, nil
		}
	}

	// 若满stream，则返回错误，让调用者创建新conn
	if len(c.streams) >= c.maxStreamCount {
		return nil, ErrExceedStreamMax
	}

	// 若不存在空闲stream，但没有满stream，则创建stream
	stream, err := c.conn.OpenStream()
	if err != nil {
		return nil, err
	}
	flag := &atomic.Value{}
	flag.Store(idle)

	si := &StreamInfo{
		stream: stream,
		flag:   flag,
	}
	c.streams = append(c.streams, si)

	return si, nil
}

// existsAvailableStream 无锁
func (c *ConnInfo) existsAvailableStream() bool {
	// 外面锁，里面也锁，当然会deadlock啊
	for _, si := range c.streams {
		if si.flag.Load().(int) == idle {
			return true
		}
	}

	return false
}

type StreamInfo struct {
	stream quic.Stream
	flag   *atomic.Value
}

type AdapterStreamConn struct {
	ci *ConnInfo
	si *StreamInfo
}

func (sc *AdapterStreamConn) Close() error {
	// 我并不认为close stream有什么用
	// Close()会和AcquireStream()冲突吗？如果去掉ci.lock的话。
	// 在先close的情况下，acquire没有得到最新，就会创建多余的stream
	// 在tryGetStream中加锁也是很有可能在间隙中创建多余的stream，但不会超过最大限制

	sc.ci.rwMutex.Lock()
	sc.si.flag.Store(idle)
	if !sc.ci.existsAvailableStream() {
		sc.ci.lastUseTime = time.Now().Add(expireDuration)
	}
	sc.ci.rwMutex.Unlock()

	return nil
}

func (sc *AdapterStreamConn) Read(p []byte) (n int, err error) {
	return sc.si.stream.Read(p)
}

func (sc *AdapterStreamConn) Write(p []byte) (n int, err error) {
	return sc.si.stream.Write(p)
}
