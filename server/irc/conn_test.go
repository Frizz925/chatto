package irc

import (
	"chatto/util/env"
	utesting "chatto/util/testing"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnection(t *testing.T) {
	if !env.IsIntegrationTest() {
		t.SkipNow()
		return
	}

	require := require.New(t)
	ctx, cancel := utesting.CreateTestingContext()
	defer cancel()

	c := NewConn(Config{
		Nick: "chatto-test",
		Name: "chatto-irc test client",
	})

	// Make sure all the registered events are called
	counter := utesting.NewCounter()
	countHandler := func(Event) {
		counter.Add()
	}
	c.Each(ctx, CONNECTED, countHandler)
	c.Each(ctx, JOIN, countHandler)
	c.Each(ctx, QUIT, countHandler)
	c.Each(ctx, DISCONNECTED, countHandler)

	require.Nil(c.Connect(ctx, "127.0.0.1:6667"))
	require.True(c.Connected())

	channel := "#chatto-test"
	require.Nil(c.Join(ctx, channel))
	require.Nil(c.Privmsg(ctx, channel, "Hello, world!"))
	require.Nil(c.Close(ctx))
	require.False(c.Connected())

	// Count the called registered events
	require.Equal(4, counter.Int())
}
