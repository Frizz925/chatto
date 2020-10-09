package testing

import (
	"chatto/util/env"
	"context"
	"time"
)

const defaultContextTimeout = 15 * time.Second

func CreateTestingContext() (context.Context, context.CancelFunc) {
	if env.IsAppMode("DEBUG") {
		return context.WithCancel(context.Background())
	}
	return context.WithTimeout(context.Background(), defaultContextTimeout)
}
