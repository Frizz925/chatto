package irc

import (
	"context"
	"net"
	"sync"
)

type Client struct {
	*Commands
	cfg       Config
	conn      net.Conn
	mu        sync.RWMutex
	connected bool
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg:       cfg,
		connected: false,
	}
}

func (c *Client) Connected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected && c.Commands.Connected()
}

func (c *Client) Connect(ctx context.Context, addr string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.connected {
		return ErrAlreadyConnected
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	cmds := NewCommands(conn, c.cfg)
	if err := cmds.Connect(ctx); err != nil {
		return err
	}
	c.conn = conn
	c.Commands = cmds
	c.connected = true
	return nil
}

func (c *Client) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.connected {
		return ErrNotConnected
	}
	if err := c.Commands.Close(ctx); err != nil {
		return err
	}
	if err := c.conn.Close(); err != nil {
		return err
	}
	c.conn = nil
	c.Commands = nil
	c.connected = false
	return nil
}
