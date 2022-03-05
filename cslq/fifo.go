package cslq

type Fifo struct {
	items []interface{}
}

func NewFifo(item ...interface{}) *Fifo {
	return &Fifo{items: item}
}

func (fifo *Fifo) Pop() interface{} {
	if len(fifo.items) == 0 {
		panic("fifo empty")
	}

	i := fifo.items[0]
	fifo.items = fifo.items[1:]

	return i
}

func (fifo *Fifo) Len() int {
	return len(fifo.items)
}
