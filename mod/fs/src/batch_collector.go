package fs

type BatchCollector[T any] struct {
	batch   []T
	size    int
	process func([]T) error
}

func NewBatchCollector[T any](size int, process func([]T) error) *BatchCollector[T] {
	return &BatchCollector[T]{
		batch:   make([]T, 0, size),
		size:    size,
		process: process,
	}
}

func (c *BatchCollector[T]) Add(item T) error {
	c.batch = append(c.batch, item)
	if len(c.batch) >= c.size {
		return c.Flush()
	}
	return nil
}

func (c *BatchCollector[T]) Flush() error {
	if len(c.batch) == 0 {
		return nil
	}
	err := c.process(c.batch)
	c.batch = c.batch[:0]
	return err
}

func (c *BatchCollector[T]) Iter(iter func(yield func(T) error) error) error {
	if err := iter(c.Add); err != nil {
		return err
	}
	return c.Flush()
}
