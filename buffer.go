package main

import (
	"fmt"
	"io"
)

type buffer struct {
	buf []byte
}

func (b *buffer) Read(p []byte) (n int, err error) {
	if len(p) > len(b.buf) {
		return copy(p, b.buf), io.EOF
	}

	return copy(p, b.buf[:len(p)]), nil
}

func (b *buffer) ReadAt(p []byte, off int) (n int, err error) {
	var bufsize = len(b.buf)
	if off < 0 || off >= bufsize {
		return 0, fmt.Errorf("offset out of buffer range")
	}

	if len(p) > bufsize-off {
		return copy(p, b.buf[off:]), io.EOF
	}

	return copy(p, b.buf[off:len(p)+off]), nil
}

func (b *buffer) Write(p []byte) (n int, err error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *buffer) WriteAt(p []byte, off int) (n int, err error) {
	var bufsize = len(b.buf)
	if off < 0 {
		return 0, fmt.Errorf("offset out of buffer range")
	}

	if off > bufsize {
		grow := make([]byte, off-bufsize)
		copy(grow, b.buf)
		b.buf = grow
	}

	required := off + len(p)
	if required > bufsize {
		grow := make([]byte, required)
		n = copy(grow, append(b.buf[:off], p...))
		b.buf = grow
		return
	}

	return copy(b.buf[off:off+len(p)], p), nil
}

func (b *buffer) Truncate(size int) error {
	if size > len(b.buf) {
		b.buf = append(b.buf, make([]byte, size-len(b.buf))...)
		return nil
	}

	b.buf = b.buf[:size]
	return nil
}

func (b *buffer) Len() int {
	return len(b.buf)
}
