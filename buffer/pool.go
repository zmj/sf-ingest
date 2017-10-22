package buffer

import "sync"

type Buffer struct {
	B    []byte
	pool *Pool
}

type Pool struct {
	mu   *sync.Mutex
	free [][]byte
	size int
}

func NewPool() *Pool {
	return &Pool{mu: &sync.Mutex{}, size: 4096}
}

func (b *Buffer) Free() {
	if b == nil {
		return
	}
	if b.pool == nil {
		b.B = nil
		return
	}
	b.pool.mu.Lock()
	defer b.pool.mu.Unlock()
	b.pool.free = append(b.pool.free, b.B)
	b.B = nil
}

func (p *Pool) GetBuffer() *Buffer {
	if p == nil {
		p = NewPool()
	}
	if len(p.free) == 0 {
		return p.new()
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.free) == 0 {
		return p.new()
	}
	b := p.free[0]
	p.free = p.free[1:]
	return &Buffer{B: b, pool: p}
}

func (p *Pool) new() *Buffer {
	return &Buffer{B: make([]byte, 0, p.size), pool: p}
}

func (p *Pool) BufferSize() int {
	return p.size
}