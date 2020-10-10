package main

import (
	"chatto/irc"
	"context"
	"os"
	"os/signal"
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
		log.Infof("Joined channel %s", cfg.Channel)
	})
	conn.Each(ctx, irc.DISCONNECTED, func(e irc.Event) {
		log.Infof("Disconnected from IRC server %s", cfg.Addr)
	})

	if err := conn.Connect(ctx, cfg.Addr); err != nil {
		return err
	}
	if err := conn.Join(ctx, cfg.Channel); err != nil {
		return err
	}
	if err := conn.Privmsg(ctx, cfg.Channel, "Hello, world!"); err != nil {
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
