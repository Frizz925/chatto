package testing

import "sync/atomic"

type Counter struct {
	value *atomic.Value
}

func NewCounter() *Counter {
	return &Counter{
		value: &atomic.Value{},
	}
}

func (c *Counter) Add() {
	x := c.Int()
	c.value.Store(x + 1)
}

func (c *Counter) Int() int {
	if x, ok := c.value.Load().(int); ok {
		return x
	}
	return 0
}
