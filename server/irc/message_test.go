package irc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessage(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Parse JOIN message
	{
		line := ":chatto!~chatto-irc@9qt4sazudxvsk.irc JOIN #chatto"
		msg := parseLine(line)
		assert.Equal(line, msg.Raw)
		assert.Equal("chatto!~chatto-irc@9qt4sazudxvsk.irc", msg.Src)
		assert.Equal("chatto", msg.Nick)
		assert.Equal("~chatto-irc", msg.Ident)
		assert.Equal("9qt4sazudxvsk.irc", msg.Host)
		assert.Equal("JOIN", msg.Cmd)
		require.GreaterOrEqual(len(msg.Args), 1)
		assert.Equal("#chatto", msg.Args[0])
	}

	// Parse ERROR message
	{
		line := "ERROR :Ping timeout: 2m30s"
		msg := parseLine(line)
		assert.Equal(line, msg.Raw)
		assert.Empty(msg.Src)
		assert.Equal("ERROR", msg.Cmd)
		require.GreaterOrEqual(len(msg.Args), 1)
		assert.Equal("Ping timeout: 2m30s", msg.Args[0])
	}

	// Parse QUIT message
	{
		line := ":chatto!~chatto-irc@9qt4sazudxvsk.irc QUIT :Ping timeout: 2m30s"
		msg := parseLine(line)
		assert.Equal(line, msg.Raw)
		assert.Equal("chatto!~chatto-irc@9qt4sazudxvsk.irc", msg.Src)
		assert.Equal("chatto", msg.Nick)
		assert.Equal("~chatto-irc", msg.Ident)
		assert.Equal("9qt4sazudxvsk.irc", msg.Host)
		assert.Equal("QUIT", msg.Cmd)
		require.GreaterOrEqual(len(msg.Args), 1)
		assert.Equal("Ping timeout: 2m30s", msg.Args[0])
	}
}
