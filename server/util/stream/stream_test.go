package stream

import (
	utesting "chatto/util/testing"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStream(t *testing.T) {
	require := require.New(t)
	ctx, cancel := utesting.CreateTestingContext()
	defer cancel()

	// Test for repeated call with same key results in the same reference
	stream := NewStream(ctx)
	require.Equal(stream.Topic(t), stream.Topic(t))
}
