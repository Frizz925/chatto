package stream

import (
	utesting "chatto/util/testing"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStream(t *testing.T) {
	require := require.New(t)
	ctx, cancel := utesting.CreateTestingContext()
	defer cancel()

	stream := NewStream()
	defer stream.Close()
	ch, id := stream.Observe(t)
	expected := "Foobar"

	// Test stream notify
	{
		timeout := time.After(1 * time.Second)
		stream.Notify(t, Item{V: expected})
		select {
		case item := <-ch:
			require.Nil(item.E)
			require.Equal(expected, item.V)
		case <-timeout:
			require.FailNow("Expected to receive event before timeout")
		}
	}

	// Test stream remove
	{
		timeout := time.After(100 * time.Millisecond)
		stream.Remove(t, id)
		stream.Notify(t, Item{})
		select {
		case <-ch:
			require.FailNow("Expected to not receive event after observer removed")
		case <-timeout:
		}
	}

	// Test stream observe func
	{
		key := 0
		timeout := time.After(1 * time.Second)
		ch := make(chan Item)
		stream.ObserveFunc(ctx, key, func(item Item) {
			ch <- item
		})
		stream.Notify(key, Item{V: expected})
		select {
		case item := <-ch:
			require.Nil(item.E)
			require.Equal(expected, item.V)
		case <-timeout:
			require.FailNow("Expected to receive event before timeout")
		}
	}
}
