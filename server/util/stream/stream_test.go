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

	// Test stream each
	{
		key := 0
		timeout := time.After(1 * time.Second)
		ch := make(chan Item)
		stream.Each(ctx, key, func(item Item) {
			ch <- item
		})
		stream.Notify(key, Item{V: 0})
		stream.Notify(key, Item{V: 1})
		assertObservableValue(require, ch, timeout, 0)
		assertObservableValue(require, ch, timeout, 1)
	}

	// Test stream once
	{
		key := 0
		timeout := time.After(1 * time.Second)
		ch := make(chan Item)
		stream.Once(ctx, key, func(item Item) {
			ch <- item
		})
		stream.Notify(key, Item{V: expected})
		stream.Notify(key, Item{})
		assertObservableValue(require, ch, timeout, expected)
		assertObservableTimeout(require, ch, timeout)
	}
}
