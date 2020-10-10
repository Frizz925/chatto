package irc

import (
	"chatto/irc"
	"context"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Handler struct {
	context context.Context
}

func New(ctx context.Context) *Handler {
	return &Handler{context: ctx}
}

func (h *Handler) Join(e irc.Event) {
	client, args := e.Client, e.Message.Args
	if len(args) < 1 {
		return
	}
	channel := args[0]
	log.Infof("Joined channel %s", channel)
	if err := client.Privmsg(h.context, channel, "Hello, world!"); err != nil {
		log.Errorf("Failed to send message to %s: %+v", channel, err)
	} else {
		log.Infof("Sent hello message to %s", channel)
	}
}

func (h *Handler) Invite(e irc.Event) {
	client, args := e.Client, e.Message.Args
	if len(args) < 2 {
		return
	}
	channel := args[1]
	log.Infof("Invited to channel %s", channel)
	if err := client.Join(h.context, channel); err != nil {
		log.Errorf("Failed to join channel %s: %+v", channel, err)
	}
}

func (h *Handler) Kick(e irc.Event) {
	args := e.Message.Args
	nargs := len(args)
	if nargs < 1 {
		return
	}
	channel := args[0]
	if nargs >= 3 {
		reason := args[2]
		log.Infof("Kicked from channel %s (%s)", channel, reason)
	} else {
		log.Infof("Kicked from channel %s", channel)
	}
}

func (h *Handler) Message(e irc.Event) {
	client := e.Client
	nick, args := e.Message.Nick, e.Message.Args
	if len(args) < 2 {
		return
	}
	channel, message := args[0], strings.Join(args[1:], " ")
	log.Infof("[%s] %s: %s", channel, nick, message)
	if !strings.HasPrefix(message, ".echo") {
		return
	}
	sidx := strings.Index(message, " ")
	reply := strings.TrimSpace(message[sidx+1:])
	if err := client.Privmsg(h.context, channel, reply); err != nil {
		log.Errorf("Failed to reply echo message to %s at %s: %+v", nick, channel, err)
	} else {
		log.Infof("Replied echo message to %s at %s", nick, channel)
	}
}
