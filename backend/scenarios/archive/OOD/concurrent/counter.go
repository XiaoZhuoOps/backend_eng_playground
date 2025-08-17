package OOD

import "sync/atomic"

type Counter struct {
	count int64
}

func (c *Counter) Add(n int) {
	/*
		In Go, the expression cnt += 1 (or cnt++) is actually composed of multiple steps:

		Read the current value of cnt.
		Add 1 to the read value.
		Write the updated value back to cnt.
	*/
	atomic.AddInt64(&c.count, int64(n))
}

func (c *Counter) Get() int64 {
	return atomic.LoadInt64(&c.count)
}
