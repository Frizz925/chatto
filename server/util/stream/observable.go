package stream

import (
	"context"
	"sync"
)

type ObserverFunc func(Item)

type Observable struct {
	mu sync.RWMutex
	wg sync.WaitGroup
	ch chan Item

	nextId    int
	observers map[int]chan<- Item
	cancel    context.CancelFunc
	open      bool
}

func NewObservable() *Observable {
	obs := &Observable{
		ch:        make(chan Item),
		nextId:    0,
		observers: make(map[int]chan<- Item),
	}
	obs.Open()
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

func (o *Observable) Once(ctx context.Context, handler ObserverFunc) *Observer {
	onceCtx, cancel := context.WithCancel(ctx)
	return o.Each(onceCtx, func(item Item) {
		defer cancel()
		handler(item)
	})
}

func (o *Observable) Each(ctx context.Context, handler ObserverFunc) *Observer {
	ch, id := o.Observe()
	loopCtx, cancel := context.WithCancel(ctx)
	observer := NewObserver(o, id, cancel)
	go func() {
		defer observer.Remove()
		for {
			select {
			case item := <-ch:
				go handler(item)
			case <-loopCtx.Done():
				return
			}
		}
	}()
	return observer
}

func (o *Observable) Notify(item Item) {
	o.ch <- item
}

func (o *Observable) Remove(id int) {
	o.mu.Lock()
	defer o.mu.Unlock()
	delete(o.observers, id)
}

func (o *Observable) Open() {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.open {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	o.cancel = cancel
	o.wg.Add(1)
	go o.notifyLoop(ctx)
	o.open = true
}

func (o *Observable) Close() {
	o.mu.Lock()
	defer o.mu.Unlock()
	if !o.open {
		return
	}
	o.cancel()
	o.wg.Wait()
	o.cancel = nil
	o.open = false
}

func (o *Observable) notifyLoop(ctx context.Context) {
	defer o.wg.Done()
	for {
		select {
		case item := <-o.ch:
			for _, ch := range o.observers {
				ch <- item
			}
		case <-ctx.Done():
			return
		}
	}
}
