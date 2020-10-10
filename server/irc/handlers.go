package irc

import (
	"context"

	log "github.com/sirupsen/logrus"
)

type intHandlerFunc func(*handlers, Event)

type handlers struct {
	context context.Context
}

var intHandlers = map[string]intHandlerFunc{
	PING: (*handlers).ping,
}

func registerInternalHandlers(ctx context.Context, c *Client) {
	h := &handlers{ctx}
	for name, handler := range intHandlers {
		c.Each(ctx, name, func(e Event) {
			handler(h, e)
		})
	}
}

func (h *handlers) ping(e Event) {
	client := e.Client
	if err := client.Pong(h.context, client.nick); err != nil {
		log.Errorf("Error replying PING: %+v", err)
	}
}
