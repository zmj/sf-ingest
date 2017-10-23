package buffer

import "sync"
import "fmt"

type Buffer struct {
	B     []byte
	array [4096]byte
	pool  *Pool
}

type Pool struct {
	mu   *sync.Mutex
	free [][4096]byte
	size int
}

func NewPool() *Pool {
	return &Pool{mu: &sync.Mutex{}, size: 4096}
}

func (b *Buffer) Free() {
	if b == nil || b.B == nil || b.pool == nil {
		return
	}
	b.pool.mu.Lock()
	defer b.pool.mu.Unlock()
	fmt.Printf("pool add %v %v\n", len(b.B), cap(b.B))
	b.pool.free = append(b.pool.free, b.array)
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
	arr := p.free[0]
	p.free = p.free[1:]
	fmt.Printf("pool get %v %v\n", len(arr), cap(arr))
	return &Buffer{B: arr[:], array: arr, pool: p}
}

func (p *Pool) new() *Buffer {
	var arr [4096]byte
	return &Buffer{B: arr[:], array: arr, pool: p}
}

func (p *Pool) BufferSize() int {
	return p.size
}
