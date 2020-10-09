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

	obs := NewObservable(ctx)
	ch, id := obs.Observe()
	expected := "Foobar"

	{
		timeout := time.After(1 * time.Second)
		obs.Notify(Item{V: expected})
		select {
		case item := <-ch:
			require.Nil(item.E)
			require.Equal(expected, item.V)
		case <-timeout:
			require.FailNow("Expected to receive event before timeout")
		}
	}

	{
		timeout := time.After(1 * time.Second)
		obs.Remove(id)
		obs.Notify(Item{})
		select {
		case <-ch:
			require.FailNow("Expected to not receive event after observer removed")
		case <-timeout:
		}
	}
}
