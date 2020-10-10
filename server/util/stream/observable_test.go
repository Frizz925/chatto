package stream

import (
	utesting "chatto/util/testing"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestObservable(t *testing.T) {
	require := require.New(t)
	ctx, cancel := utesting.CreateTestingContext()
	defer cancel()

	expected := "Foobar"

	// Test observable notify and remove
	{
		obs := NewObservable()
		defer obs.Close()
		ch, id := obs.Observe()
		timeout := time.After(1 * time.Second)

		obs.Notify(Item{V: expected})
		assertObservableValue(require, ch, timeout, expected)

		obs.Remove(id)
		obs.Notify(Item{})
		assertObservableTimeout(require, ch, timeout)
	}

	// Test observable each
	{
		obs := NewObservable()
		defer obs.Close()
		timeout := time.After(1 * time.Second)
		ch := make(chan Item)
		obs.Each(ctx, func(item Item) {
			ch <- item
		})
		go func() {
			obs.Notify(Item{V: 0})
			obs.Notify(Item{V: 1})
		}()
		assertObservableValue(require, ch, timeout, 0)
		assertObservableValue(require, ch, timeout, 1)
	}

	// Test observable once
	{
		obs := NewObservable()
		defer obs.Close()
		timeout := time.After(1 * time.Second)
		ch := make(chan Item)
		obs.Once(ctx, func(item Item) {
			ch <- item
		})
		go func() {
			obs.Notify(Item{V: expected})
			obs.Notify(Item{})
		}()
		assertObservableValue(require, ch, timeout, expected)
		assertObservableTimeout(require, ch, timeout)
	}
}

func assertObservableValue(
	require *require.Assertions,
	ch <-chan Item,
	timeout <-chan time.Time,
	expected interface{},
) {
	select {
	case item := <-ch:
		require.Nil(item.E)
		require.Equal(expected, item.V)
	case <-timeout:
		require.FailNow("Expected to receive event before timeout")
	}
}

func assertObservableTimeout(
	require *require.Assertions,
	ch <-chan Item,
	timeout <-chan time.Time,
) {
	select {
	case <-ch:
		require.FailNow("Expected NOT to receive event before timeout")
	case <-timeout:
	}
}
