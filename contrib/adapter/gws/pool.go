package gws

import (
	"bytes"
	"sync"
)

type writerPool struct {
	p sync.Pool
}

func newWriterPool() *writerPool {
	return &writerPool{
		p: sync.Pool{
			New: func() any {
				return &responseWriter{
					buf:      &bytes.Buffer{},
					payloads: make([][]byte, 1, 2),
				}
			},
		},
	}
}

func (c *writerPool) Get() *responseWriter {
	return c.p.Get().(*responseWriter)
}

func (c *writerPool) Put(w *responseWriter) {
	w.conn = nil
	w.headerCodec = nil
	w.code = 0
	w.buf.Reset()
	w.payloads = w.payloads[:0]
	c.p.Put(w)
}
