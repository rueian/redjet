package redjet

import (
	"context"
	"net"
	"time"
)

type conn struct {
	net.Conn

	lastUsed time.Time
}

type connPool struct {
	free        chan *conn
	canceled    chan struct{}
	cleanTicker *time.Ticker
	idleTimeout time.Duration
}

func newConnPool(size int, idleTimeout time.Duration) *connPool {
	p := &connPool{
		free: make(chan *conn, size),
		// 3 is chosen arbitrarily.
		cleanTicker: time.NewTicker(idleTimeout * 3),
		canceled:    make(chan struct{}),
		idleTimeout: idleTimeout,
	}
	go p.clean()
	return p
}

func (p *connPool) clean() {
	// We use a centralized routine for cleaning instead of AfterFunc on each
	// connection because the latter creates more garbage, even though it scales
	// logarithmically as opposed to linearly.
	for {
		select {
		case <-p.canceled:
			return
		case <-p.cleanTicker.C:
			for {
				select {
				// Remove all idle connections.
				case c := <-p.free:
					if time.Since(c.lastUsed) > p.idleTimeout {
						c.Close()
						continue
					}
					p.free <- c
				default:
					return
				}
			}
		}
	}
}

// tryGet tries to get a connection from the pool. If there are no free
// connections, it returns false.
func (p *connPool) tryGet(ctx context.Context) (*conn, bool) {
	select {
	case c := <-p.free:
		c.lastUsed = time.Now()
		return c, true
	default:
		return nil, false
	}
}

// put returns a connection to the pool.
// If the pool is full, the connection is closed.
func (p *connPool) put(nc net.Conn, idleTimeout time.Duration) {
	c := &conn{
		Conn: nc,
	}

	select {
	case p.free <- c:
	default:
		// Pool is full, just close the connection.
		c.Close()
	}
}