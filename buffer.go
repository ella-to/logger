package logger

import "time"

type buffer[T any] struct {
	buffer   []T
	recordCh chan T
	closeCh  chan struct{}
}

func (c *buffer[T]) Add(r T) {
	c.recordCh <- r
}

func (c *buffer[T]) Close() {
	close(c.closeCh)
}

func newBuffer[T any](size int, interval time.Duration, fn func([]T) error) *buffer[T] {
	c := &buffer[T]{
		buffer:   make([]T, 0, size),
		recordCh: make(chan T, size),
		closeCh:  make(chan struct{}),
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case r, ok := <-c.recordCh:
				if !ok {
					return
				}

				// the first condition is to be called when the buffer is already full
				// and the previous flush call failed. This is to prevent the buffer to
				// grow indefinitely. At this point, we start to lose logs.
				if len(c.buffer) >= size {
					err := fn(c.buffer)
					if err != nil {
						continue
					}
					c.buffer = c.buffer[:0]
				}

				c.buffer = append(c.buffer, r)

				if len(c.buffer) >= size {
					err := fn(c.buffer)
					if err != nil {
						continue
					}
					c.buffer = c.buffer[:0]
				}

			case <-ticker.C:
				if len(c.buffer) > 0 {
					err := fn(c.buffer)
					if err != nil {
						continue
					}

					c.buffer = c.buffer[:0]
				}

			case <-c.closeCh:
				return
			}
		}
	}()

	return c
}
