package irc

import "chatto/util/stream"

type Event struct {
	Client   *Client
	Observer *stream.Observer
	Message  Message
	Error    error
}
