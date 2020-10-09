package irc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
)

var (
	ErrAlreadyConnected = errors.New("already connected")
	ErrNotConnected     = errors.New("not connected")
)

type Config struct {
	Nick  string
	Ident string
	Name  string
}

type Commands struct {
	rw  io.ReadWriter
	cfg Config

	in     chan string
	out    chan string
	cancel context.CancelFunc

	mu sync.RWMutex
	wg sync.WaitGroup

	connected bool
	lastError error
}

func NewCommands(rw io.ReadWriter, cfg Config) *Commands {
	if cfg.Ident == "" {
		cfg.Ident = "chatto-irc"
	}
	if cfg.Name == "" {
		cfg.Name = "Chatto IRC client"
	}
	return &Commands{
		rw:        rw,
		cfg:       cfg,
		connected: false,
	}
}

func (c *Commands) Connected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

func (c *Commands) Connect(ctx context.Context) error {
	if err := c.init(ctx); err != nil {
		return err
	}
	if err := c.Nick(ctx, c.cfg.Nick); err != nil {
		return err
	}
	return c.User(ctx, c.cfg.Ident, c.cfg.Name)
}

func (c *Commands) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.connected {
		return ErrNotConnected
	}

	closeCtx, cancel := context.WithCancel(ctx)
	go func() {
		c.wg.Wait()
		cancel()
	}()
	c.cancel()

	// Wait for either the goroutines to finish or parent context being cancelled
	select {
	case <-closeCtx.Done():
		c.connected = false
		return nil
	case <-ctx.Done():
		c.connected = false
		return ctx.Err()
	}
}

func (c *Commands) Nick(ctx context.Context, nick string) error {
	return c.Command(ctx, "NICK", nick)
}

func (c *Commands) User(ctx context.Context, ident string, name string) error {
	return c.Command(ctx, "USER", ident, "12 * :"+name)
}

func (c *Commands) Join(ctx context.Context, channel string, key ...string) error {
	params := channel
	if len(key) > 0 {
		params = fmt.Sprintf("%s %s", channel, key[0])
	}
	return c.Command(ctx, "JOIN", params)
}

func (c *Commands) Privmsg(ctx context.Context, target string, msg string) error {
	return c.Command(ctx, "PRIVMSG", target, ":"+msg)
}

func (c *Commands) Command(ctx context.Context, name string, params ...string) error {
	line := strings.Join(append([]string{name}, params...), " ")
	return c.Write(ctx, line)
}

func (c *Commands) Read(ctx context.Context) (string, error) {
	if !c.Connected() {
		return "", ErrNotConnected
	}
	select {
	case line := <-c.in:
		if line != "" {
			return line, nil
		}
	case <-ctx.Done():
		return "", ctx.Err()
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return "", c.lastError
}

func (c *Commands) Write(ctx context.Context, line string) error {
	if !c.Connected() {
		return ErrNotConnected
	}
	parts := strings.SplitN(line, "\r\n", 2)
	select {
	case c.out <- parts[0] + "\r\n":
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Commands) init(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.connected {
		return ErrAlreadyConnected
	}

	loopCtx, cancel := context.WithCancel(ctx)

	c.in = make(chan string)
	c.out = make(chan string)
	c.cancel = cancel
	c.lastError = nil

	c.wg.Add(2)
	go c.readLoop(loopCtx)
	go c.writeLoop(loopCtx)

	c.connected = true
	return nil
}

func (c *Commands) readLoop(ctx context.Context) {
	defer c.wg.Done()
	reader := bufio.NewReader(c.rw)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			c.handleError(err)
			c.in <- ""
			return
		}
		select {
		case c.in <- line:
		case <-ctx.Done():
			return
		}
	}
}

func (c *Commands) writeLoop(ctx context.Context) {
	defer c.wg.Done()
	writer := bufio.NewWriter(c.rw)
	for {
		select {
		case payload := <-c.out:
			if _, err := writer.WriteString(payload); err != nil {
				c.handleError(err)
				return
			}
			if err := writer.Flush(); err != nil {
				c.handleError(err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Commands) handleError(err error) {
	if !errors.Is(err, io.EOF) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.lastError = err
	}
}
