package irc

import (
	"strings"
	"time"
)

type Message struct {
	Nick, Ident, Host, Src string
	Raw, Cmd               string
	Args                   []string
	Time                   time.Time
}

func parseLine(s string) (msg Message) {
	msg.Time = time.Now()

	if s == "" {
		return msg
	}
	msg.Raw = s

	if s[0] == ':' {
		if idx := strings.Index(s, " "); idx != -1 {
			msg.Src, s = s[1:idx], s[idx:]
		} else {
			return msg
		}
		nidx, iidx := strings.Index(msg.Src, "!"), strings.Index(msg.Src, "@")
		if nidx != -1 && iidx != -1 {
			msg.Nick = msg.Src[:nidx]
			msg.Ident = msg.Src[nidx+1 : iidx]
			msg.Host = msg.Src[iidx+1:]
		}
	}

	args := strings.SplitN(s, " :", 2)
	if len(args) > 1 {
		args = append(strings.Fields(args[0]), args[1])
	} else {
		args = strings.Fields(args[0])
	}

	nargs := len(args)
	if nargs < 1 {
		return msg
	}
	msg.Cmd = args[0]
	if nargs > 1 {
		msg.Args = args[1:]
	}
	return msg
}
