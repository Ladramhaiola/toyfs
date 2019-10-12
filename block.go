package main

import (
	"errors"
)

// Block - simple data block
type Block struct {
	size int
	data []byte
}

var (
	// ErrWriteBytes -
	ErrWriteBytes = errors.New("Not enough space to write")
	// ErrReadBytes -
	ErrReadBytes = errors.New("Read error")
	// ErrOffsetRange -
	ErrOffsetRange = errors.New("Offset out of block range")
)

// Write - write bytes to data block
func (b *Block) Write(p []byte) {
	b.data = make([]byte, blockSize)
	b.size = copy(b.data, p)
}

// WriteAt - write bytes with offset
func (b *Block) WriteAt(p []byte, off int) error {
	if off > blockSize || off < 0 {
		return ErrOffsetRange
	}

	var size = len(p)
	if off+size > blockSize {
		return ErrWriteBytes
	}

	if len(b.data) == 0 {
		b.data = make([]byte, blockSize)
	}

	counter := 0
	for i := off; i < blockSize && counter < size; i++ {
		b.data[i] = p[counter]
		counter++
	}
	b.size += counter

	return nil
}

// Read - read all block data
func (b *Block) Read() []byte {
	return b.data
}

// ReadAt - read block data with offset
func (b *Block) ReadAt(off int) ([]byte, error) {
	if off > blockSize || off < 0 {
		return []byte{}, ErrOffsetRange
	}

	return b.data[off:], nil
}

// Busy -
func (b *Block) Busy() bool {
	return len(b.data) > 0
}

// Avaivable -
func (b *Block) Avaivable() int {
	return blockSize - b.size
}
