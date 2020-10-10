package stream

import "context"

type Observer struct {
	observable *Observable
	id         int
	cancel     context.CancelFunc
}

func NewObserver(o *Observable, id int, cancel context.CancelFunc) *Observer {
	return &Observer{
		observable: o,
		id:         id,
		cancel:     cancel,
	}
}

func (o *Observer) Remove() {
	o.cancel()
	o.observable.Remove(o.id)
}
