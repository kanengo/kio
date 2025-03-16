package kio

import (
	"log/slog"
	"math/rand/v2"
	"time"

	"runtime"
)

type Engine struct {
	Config

	eventHandlers map[string]EventHandler
	mainLoop      *eventLoop
	loops         []*eventLoop

	rd *rand.Rand
}

type EventHandler interface {
	Name() string
	OnOpen(c *Conn)
	OnClose(c *Conn)
	OnRead(c *Conn, data []byte)
	OnData(c *Conn, data []byte)
}

//go:norace
func NewEngine(config Config) *Engine {
	if config.PollerNum <= 0 {
		config.PollerNum = runtime.NumCPU()
	}
	if config.ThreadPoolNum <= 0 {
		config.ThreadPoolNum = runtime.NumCPU()
	}
	if config.logger == nil {
		config.logger = slog.Default()
	}

	rd := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))
	return &Engine{
		Config: config,
		rd:     rd,
	}
}

//go:norace
func (e *Engine) RegisterEventHandler(handler EventHandler) {
	e.eventHandlers[handler.Name()] = handler
}

func (e *Engine) addConn(conn *Conn) error {
	//p2c picker
	var l1, l2 *eventLoop
	if len(e.loops) == 1 {
		e.loops[0].addConn(conn)
		return nil
	} else if len(e.loops) == 2 {
		l1 = e.loops[0]
		l2 = e.loops[1]
	} else {
		for {
			l1 = e.loops[e.rd.IntN(len(e.loops))]
			l2 = e.loops[e.rd.IntN(len(e.loops))]
			if l1 != l2 {
				break
			}
		}
	}

	c1 := l1.connNum.Load()
	c2 := l2.connNum.Load()
	if c1 < c2 {
		return l1.addConn(conn)
	} else {
		return l2.addConn(conn)
	}
}

//go:norace
func (e *Engine) Start() error {
	return nil
}
