package stream

import (
	"context"
	"sync"
)

type Observable struct {
	mu        sync.RWMutex
	ch        chan Item
	nextId    int
	observers map[int]chan<- Item
}

func NewObservable(ctx context.Context) *Observable {
	obs := &Observable{
		ch:        make(chan Item),
		nextId:    0,
		observers: make(map[int]chan<- Item),
	}
	go obs.notifyLoop(ctx)
	return obs
}

func (o *Observable) Observe() (<-chan Item, int) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.nextId++
	id := o.nextId
	ch := make(chan Item, 1)
	o.observers[id] = ch
	return ch, id
}

func (o *Observable) Remove(id int) {
	o.mu.Lock()
	defer o.mu.Unlock()
	delete(o.observers, id)
}

func (o *Observable) Notify(item Item) {
	o.ch <- item
}

func (o *Observable) notifyLoop(ctx context.Context) {
	for {
		select {
		case item := <-o.ch:
			for _, ch := range o.observers {
				select {
				case ch <- item:
				default:
				}
			}
		case <-ctx.Done():
		}
	}
}
