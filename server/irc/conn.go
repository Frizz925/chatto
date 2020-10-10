package irc

import (
	"context"
	"net"
	"sync"
)

type Conn struct {
	*Client
	conn      net.Conn
	mu        sync.RWMutex
	connected bool
}

func NewConn(cfg Config) *Conn {
	return &Conn{
		Client:    NewClient(cfg),
		connected: false,
	}
}

func (c *Conn) Connected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected && c.Client.Connected()
}

func (c *Conn) Connect(ctx context.Context, addr string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.connected {
		return ErrAlreadyConnected
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	if err := c.Client.Connect(ctx, conn); err != nil {
		return err
	}
	c.conn = conn
	c.connected = true
	return nil
}

func (c *Conn) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.connected {
		return ErrNotConnected
	}
	if err := c.Client.Close(ctx); err != nil {
		return err
	}
	if err := c.conn.Close(); err != nil {
		return err
	}
	c.conn = nil
	c.Client = nil
	c.connected = false
	return nil
}
