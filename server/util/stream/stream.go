package stream

import (
	"context"
	"sync"
)

type Stream struct {
	ctx    context.Context
	mu     sync.Mutex
	topics map[interface{}]*Observable
}

func NewStream(ctx context.Context) *Stream {
	return &Stream{
		ctx:    ctx,
		topics: make(map[interface{}]*Observable),
	}
}

func (s *Stream) Topic(key interface{}) *Observable {
	s.mu.Lock()
	defer s.mu.Unlock()
	obs, ok := s.topics[key]
	if !ok {
		obs = NewObservable(s.ctx)
		s.topics[key] = obs
	}
	return obs
}
