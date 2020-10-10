package stream

import (
	"context"
	"sync"
)

type Stream struct {
	mu     sync.RWMutex
	topics map[interface{}]*Observable
}

func NewStream() *Stream {
	return &Stream{
		topics: make(map[interface{}]*Observable),
	}
}

func (s *Stream) Observe(key interface{}) (<-chan Item, int) {
	return s.observable(key).Observe()
}

func (s *Stream) Once(ctx context.Context, key interface{}, handler ObserverFunc) *Observer {
	return s.observable(key).Once(ctx, handler)
}

func (s *Stream) Each(ctx context.Context, key interface{}, handler ObserverFunc) *Observer {
	return s.observable(key).Each(ctx, handler)
}

func (s *Stream) Notify(key interface{}, item Item) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	obs, ok := s.topics[key]
	if ok {
		obs.Notify(item)
	}
}

func (s *Stream) Remove(key interface{}, id int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	obs, ok := s.topics[key]
	if ok {
		obs.Remove(id)
	}
}

func (s *Stream) Open() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, obs := range s.topics {
		obs.Open()
	}
}

func (s *Stream) Close() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, obs := range s.topics {
		obs.Close()
	}
}

func (s *Stream) observable(key interface{}) *Observable {
	s.mu.Lock()
	defer s.mu.Unlock()
	obs, ok := s.topics[key]
	if !ok {
		obs = NewObservable()
		s.topics[key] = obs
	}
	return obs
}
