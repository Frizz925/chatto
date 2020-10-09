package main

import (
	"chatto/irc"
	"context"
	"os"
	"os/signal"
	"syscall"

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
	c := irc.NewClient(irc.Config{
		Nick: cfg.Nick,
	})

	if err := c.Connect(ctx, cfg.Addr); err != nil {
		return err
	}
	defer c.Close(ctx)
	log.Infof("Connected to IRC server %s", cfg.Addr)

	if err := c.Join(ctx, cfg.Channel); err != nil {
		return err
	}
	log.Infof("Joined channel %s", cfg.Channel)

	for {
		line, err := c.Read(ctx)
		if err != nil {
			switch err {
			case context.Canceled:
				return nil
			case context.DeadlineExceeded:
				return nil
			}
			return err
		}
		log.Info(line)
	}
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
}
