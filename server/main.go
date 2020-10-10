package main

import (
	"chatto/irc"
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	Addr    string
	Channel string
	Nick    string
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-ch
		log.Infof("Received signal %+v", sig)
		cancel()
	}()
	err := serve(ctx, Config{
		Addr:    "localhost:6667",
		Channel: "#chatto",
		Nick:    "chatto",
	})
	if err != nil {
		log.Fatal(err)
	}
}

func serve(ctx context.Context, cfg Config) error {
	conn := irc.NewConn(irc.Config{
		Nick: cfg.Nick,
	})

	conn.Each(ctx, irc.CONNECTED, func(e irc.Event) {
		log.Infof("Connected to IRC server %s", cfg.Addr)
	})
	conn.Each(ctx, irc.JOIN, func(e irc.Event) {
		args := e.Message.Args
		if len(args) < 1 {
			return
		}
		channel := args[0]
		log.Infof("Joined channel %s", channel)
		if err := conn.Privmsg(ctx, channel, "Hello, world!"); err != nil {
			log.Errorf("Failed to send message to %s: %+v", channel, err)
		} else {
			log.Infof("Sent hello message to %s", channel)
		}
	})
	conn.Each(ctx, irc.INVITE, func(e irc.Event) {
		args := e.Message.Args
		if len(args) < 2 {
			return
		}
		channel := args[1]
		log.Infof("Invited to channel %s", channel)
		if err := conn.Join(ctx, channel); err != nil {
			log.Errorf("Failed to join channel %s: %+v", channel, err)
		}
	})
	conn.Each(ctx, irc.KICK, func(e irc.Event) {
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
	})
	conn.Each(ctx, irc.PRIVMSG, func(e irc.Event) {
		nick, args := e.Message.Nick, e.Message.Args
		if len(args) < 2 {
			return
		}
		channel, message := args[0], strings.Join(args[1:], " ")
		log.Infof("[%s] %s: %s", channel, nick, message)
	})
	conn.Each(ctx, irc.DISCONNECTED, func(e irc.Event) {
		log.Infof("Disconnected from IRC server %s", cfg.Addr)
	})

	if err := conn.Connect(ctx, cfg.Addr); err != nil {
		return err
	}

	<-ctx.Done()

	log.Info("Terminating connection...")
	closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := conn.Close(closeCtx); err != nil {
		return err
	}
	log.Info("Connection terminated.")

	return nil
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
}
