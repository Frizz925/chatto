package irc

import "chatto/util/stream"

type Event struct {
	Client  *Client
	Message Message
	Error   error
}

func eventFromStream(client *Client, item stream.Item) Event {
	event := Event{
		Client: client,
		Error:  item.E,
	}
	if message, ok := item.V.(Message); ok {
		event.Message = message
	}
	return event
}
