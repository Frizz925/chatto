package irc

import (
	"chatto/util/env"
	utesting "chatto/util/testing"
	"testing"
	"time"

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

	c := NewClient(Config{
		Nick: "chatto-test",
		Name: "chatto-irc test client",
	})
	require.Nil(c.Connect(ctx, "127.0.0.1:6667"))
	require.True(c.Connected())
	defer c.Close(ctx)

	channel := "#chatto-test"
	require.Nil(c.Join(ctx, channel))
	require.Nil(c.Privmsg(ctx, channel, "Hello, world!"))

	time.Sleep(100 * time.Millisecond)
}
