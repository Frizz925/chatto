package irc

import (
	"bufio"
	"chatto/util/stream"
	"context"
	"errors"
	"io"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

var (
	ErrAlreadyConnected = errors.New("already connected")
	ErrNotConnected     = errors.New("not connected")
)

type HandlerFunc func(Event)

type Config struct {
	Nick  string
	Ident string
	Name  string
}

type Client struct {
	*Commands
	stream *stream.Stream

	cfg Config

	out    chan string
	cancel context.CancelFunc

	mu sync.RWMutex
	wg sync.WaitGroup

	connected bool
	lastError error
}

func NewClient(cfg Config) *Client {
	if cfg.Ident == "" {
		cfg.Ident = "chatto-irc"
	}
	if cfg.Name == "" {
		cfg.Name = "Chatto IRC client"
	}
	stream := stream.NewStream()
	out := make(chan string)
	return &Client{
		Commands:  NewCommands(stream, out),
		stream:    stream,
		cfg:       cfg,
		out:       out,
		connected: false,
	}
}

func (c *Client) Connected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

func (c *Client) Connect(ctx context.Context, rw io.ReadWriter) error {
	if err := c.init(rw); err != nil {
		return err
	}
	c.stream.Open()
	if err := c.Nick(ctx, c.cfg.Nick); err != nil {
		return err
	}
	if err := c.User(ctx, c.cfg.Ident, c.cfg.Name); err != nil {
		return err
	}
	c.notify(CONNECTED)
	return nil
}

func (c *Client) Close(ctx context.Context) error {
	if !c.Connected() {
		return ErrNotConnected
	}
	if err := c.Quit(ctx); err != nil {
		return err
	}
	if err := c.terminate(ctx); err != nil {
		return err
	}
	c.notify(DISCONNECTED)
	c.stream.Close()
	return nil
}

func (c *Client) Once(ctx context.Context, name string, handler HandlerFunc) *stream.Observer {
	return c.stream.Once(ctx, name, func(item stream.Item) {
		handler(eventFromStream(c, item))
	})
}

func (c *Client) Each(ctx context.Context, name string, handler HandlerFunc) *stream.Observer {
	return c.stream.Each(ctx, name, func(item stream.Item) {
		handler(eventFromStream(c, item))
	})
}

func (c *Client) Remove(name string, id int) {
	c.stream.Remove(name, id)
}

func (c *Client) init(rw io.ReadWriter) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.connected {
		return ErrAlreadyConnected
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	c.lastError = nil

	c.wg.Add(2)
	go c.recv(ctx, rw)
	go c.send(ctx, rw)

	c.Each(ctx, PING, func(e Event) {
		if err := e.Client.Pong(ctx, c.cfg.Nick); err != nil {
			log.Errorf("Error replying PING: %+v", err)
		}
	})

	c.connected = true
	return nil
}

func (c *Client) terminate(ctx context.Context) (err error) {
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
	case <-ctx.Done():
		err = ctx.Err()
	}

	c.connected = false
	return err
}

func (c *Client) recv(ctx context.Context, r io.Reader) {
	defer c.wg.Done()
	reader := bufio.NewReader(r)
	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			c.handleError(err)
			return
		}
		c.handleLine(strings.Trim(s, "\r\n"))
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func (c *Client) send(ctx context.Context, w io.Writer) {
	defer c.wg.Done()
	writer := bufio.NewWriter(w)
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

func (c *Client) handleLine(line string) {
	msg := parseLine(line)
	if msg.Cmd != "" {
		c.notify(msg.Cmd, msg)
	}
}

func (c *Client) handleError(err error) {
	if !errors.Is(err, io.EOF) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.lastError = err
	}
}

func (c *Client) notify(name string, messages ...Message) {
	if len(messages) <= 0 {
		c.stream.Notify(name, stream.Item{})
		return
	}
	for _, message := range messages {
		c.stream.Notify(name, stream.Item{V: message})
	}
}
