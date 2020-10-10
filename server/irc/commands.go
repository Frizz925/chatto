package irc

import (
	"chatto/util/stream"
	"context"
	"fmt"
	"strings"
)

const (
	CONNECTED    = "CONNECTED"
	DISCONNECTED = "DISCONNECTED"
	PRIVMSG      = "PRIVMSG"
	NICK         = "NICK"
	USER         = "USER"
	JOIN         = "JOIN"
	INVITE       = "INVITE"
	PART         = "PART"
	KICK         = "KICK"
	PING         = "PING"
	PONG         = "PONG"
	QUIT         = "QUIT"
	RAW          = "RAW"
)

type Commands struct {
	stream *stream.Stream
	out    chan<- string
}

func NewCommands(stream *stream.Stream, out chan<- string) *Commands {
	return &Commands{stream, out}
}

func (c *Commands) Nick(ctx context.Context, nick string) error {
	return c.Command(ctx, NICK, nick)
}

func (c *Commands) User(ctx context.Context, ident string, name string) error {
	return c.Command(ctx, USER, ident, "12 * :"+name)
}

func (c *Commands) Join(ctx context.Context, channel string, key ...string) error {
	params := channel
	if len(key) > 0 {
		params = fmt.Sprintf("%s %s", channel, key[0])
	}
	_, err := c.CommandWait(ctx, JOIN, params)
	return err
}

func (c *Commands) Privmsg(ctx context.Context, target string, msg string) error {
	return c.Command(ctx, PRIVMSG, target, ":"+msg)
}

func (c *Commands) Ping(ctx context.Context, dst string) error {
	return c.Command(ctx, PING, dst)
}

func (c *Commands) Pong(ctx context.Context, src string) error {
	return c.Command(ctx, PONG, src)
}

func (c *Commands) Quit(ctx context.Context, messages ...string) error {
	if len(messages) <= 0 {
		return c.Command(ctx, QUIT)
	}
	_, err := c.CommandWait(ctx, QUIT, ":"+strings.Join(messages, " "))
	return err
}

func (c *Commands) CommandWait(ctx context.Context, cmd string, args ...string) (Message, error) {
	ch, id := c.stream.Observe(cmd)
	defer c.stream.Remove(cmd, id)
	if err := c.Command(ctx, cmd, args...); err != nil {
		return Message{}, err
	}
	select {
	case item := <-ch:
		if item.E != nil {
			return Message{}, item.E
		}
		if message, ok := item.V.(Message); ok {
			return message, nil
		}
		return Message{}, nil
	case <-ctx.Done():
		return Message{}, ctx.Err()
	}
}

func (c *Commands) Command(ctx context.Context, cmd string, args ...string) error {
	line := strings.Join(append([]string{cmd}, args...), " ")
	return c.Write(ctx, line)
}

func (c *Commands) Write(ctx context.Context, line string) error {
	parts := strings.SplitN(line, "\r\n", 2)
	select {
	case c.out <- parts[0] + "\r\n":
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
