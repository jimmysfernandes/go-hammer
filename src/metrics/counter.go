package metrics

import "sync/atomic"

type Counter struct {
	count atomic.Int64
}

func (c *Counter) Inc() {
	c.count.Add(1)
}

func (c *Counter) Value() int64 {
	return c.count.Load()
}

func (c *Counter) Reset() {
	c.count.Store(0)
}

func NewCounter() Counter {
	return Counter{}
}
