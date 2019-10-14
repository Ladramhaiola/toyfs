package memfs

import (
	"bytes"
	"os"
	"path/filepath"
	"time"

	vfs "fs"
)

// File in-memory representation
type File struct {
	id      uint64
	name    string
	dir     bool
	mode    os.FileMode
	parent  *File
	linked  string
	size    int64
	modtime time.Time
	fs      vfs.Filesystem
	childs  map[string]*File
	links   []string
	data    []*Block
}

// Sys returns underlying data source
func (f *File) Sys() interface{} {
	return f.fs
}

// IsDir - check if dir
func (f *File) IsDir() bool {
	return f.dir
}

// ModTime returns last modification time
func (f *File) ModTime() time.Time {
	return f.modtime
}

// Mode - to implemet interface
func (f *File) Mode() os.FileMode {
	return f.mode
}

// Name of the file
func (f *File) Name() string {
	return f.name
}

// ID of the file
func (f *File) ID() uint64 {
	return f.id
}

// AbsPath - absolute path to file
func (f *File) AbsPath() string {
	if f.parent != nil {
		return filepath.Join(f.parent.AbsPath(), f.name)
	}
	return "/"
}

// Write - append data to File
func (f *File) Write(p []byte) (int, error) {
	var buffer = bytes.NewBuffer(p)
	var left = len(p)

	// if last already existing data block have some avaivable space
	if len(f.data) > 1 {
		tail := f.data[len(f.data)-1]
		free := tail.Avaivable()
		tail.WriteAt(buffer.Next(free), blockSize-free)
		left -= free
	}

	// required block count to write all data
	required := left / blockSize
	if left%blockSize != 0 {
		required++
	}

	for i := 0; i < required; i++ {
		block := &Block{}
		block.Write(buffer.Next(blockSize))
		f.data = append(f.data, block)
	}

	return len(p), nil
}

// WriteAt - write data with offset
func (f *File) WriteAt(p []byte, off int) (int, error) {
	var buffer = bytes.NewBuffer(p)
	var left = len(p)

	blockOffset := off / blockSize
	bytesOffset := off % blockSize

	// fill File with empty data if offset higher than File size
	if len(f.data) <= blockOffset {
		var diff = blockOffset - len(f.data)
		for i := 0; i <= diff; i++ {
			f.data = append(f.data, &Block{data: make([]byte, blockSize)})
		}
	}

	head := f.data[blockOffset]
	head.WriteAt(buffer.Next(blockSize-bytesOffset), bytesOffset)
	left -= blockSize - bytesOffset

	// required block count to write rest of the data
	required := left / blockSize
	if left%blockSize != 0 {
		required++
	}

	// append additional data to File if needed or rewrite existing
	if len(f.data) <= blockOffset+required {
		var diff = blockOffset + required - len(f.data)

		for i := 0; i <= diff; i++ {
			block := &Block{}
			block.Write(buffer.Next(blockSize))
			f.data = append(f.data, block)
		}
	} else {
		for i := 1; i <= required; i++ {
			f.data[blockOffset+i].Write(buffer.Next(blockSize))
		}
	}

	return len(p), nil
}

// Read - read all File data
func (f *File) Read() []byte {
	var buffer = new(bytes.Buffer)
	for _, block := range f.data {
		buffer.Write(block.Read())
	}
	return buffer.Bytes()
}

// ReadAt - read File data with offset
func (f *File) ReadAt(p []byte, off int) (int, error) {
	var buffer = new(bytes.Buffer)

	blockOffset := off / blockSize
	bytesOffset := off % blockSize

	if len(f.data) <= blockOffset {
		return 0, ErrOffsetRange
	}

	head, err := f.data[blockOffset].ReadAt(bytesOffset)
	if err != nil {
		return 0, ErrReadBytes
	}
	buffer.Write(head)

	for i := blockOffset + 1; i < len(f.data); i++ {
		buffer.Write(f.data[i].Read())
	}

	return buffer.Read(p)
}

// Truncate - change File size
func (f *File) Truncate(size int) error {
	blockCount := size / blockSize
	if size%blockSize != 0 {
		blockCount++
	}

	if len(f.data) > blockCount {
		f.data = f.data[:blockCount]
		return nil
	}

	var diff = blockCount - len(f.data)
	for i := 0; i <= diff; i++ {
		f.data = append(f.data, &Block{})
	}

	return nil
}

// Size in bytes
func (f *File) Size() int64 {
	if f.dir {
		return 0
	}
	return int64(len(f.data) * blockSize)
}

// Blocks count
func (f *File) Blocks() int {
	return len(f.data)
}

// Close the file
func (f *File) Close() error {
	return nil
}
