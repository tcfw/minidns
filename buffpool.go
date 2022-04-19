package main

import (
	"bytes"
	"sync"
)

// BufferPool implements a pool of bytes.Buffers in the form of a bounded
// channel.
type BufferPool struct {
	pool sync.Pool
}

// NewBufferPool creates a new BufferPool
func NewBufferPool() *BufferPool {
	var bp BufferPool
	bp.pool.New = func() interface{} {
		return new(bytes.Buffer)
	}
	return &bp
}

// Get gets a Buffer from the BufferPool, or creates a new one if none are
// available in the pool.
func (bp *BufferPool) Get() *bytes.Buffer {
	return bp.pool.Get().(*bytes.Buffer)
}

// Put returns the given Buffer to the BufferPool.
func (bp *BufferPool) Put(b *bytes.Buffer) {
	b.Reset()
	bp.pool.Put(b)
}
