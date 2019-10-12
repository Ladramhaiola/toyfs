package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// File - simple file implementation
type File struct {
	name   string
	id     uint64
	links  []string
	blocks []*Block

	dir     bool
	memDir  *DirMap
	closed  bool
	mode    os.FileMode
	modtime time.Time
}

// CreateFile -
func CreateFile(name string) *File {
	return &File{name: name, mode: os.ModeAppend, modtime: time.Now()}
}

// CreateDir -
func CreateDir(name string) *File {
	return &File{name: name, mode: os.ModeAppend, memDir: &DirMap{}, dir: true, modtime: time.Now()}
}

// SetID -
func SetID(f *File, id uint64) {
	f.id = id
}

// Write - append data to file
func (f *File) Write(p []byte) (int, error) {
	var buffer = bytes.NewBuffer(p)
	var left = len(p)

	// if last already existing data block have some avaivable space
	if len(f.blocks) > 1 {
		tail := f.blocks[len(f.blocks)-1]
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
		f.blocks = append(f.blocks, block)
	}

	f.modtime = time.Now()
	return len(p), nil
}

// WriteAt -
func (f *File) WriteAt(p []byte, off int) (int, error) {
	var buffer = bytes.NewBuffer(p)
	var left = len(p)

	blockOffset := off / blockSize
	bytesOffset := off % blockSize

	// fill file with empty blocks if offset higher than file size
	if len(f.blocks) <= blockOffset {
		var diff = blockOffset - len(f.blocks)
		for i := 0; i <= diff; i++ {
			f.blocks = append(f.blocks, &Block{data: make([]byte, blockSize)})
		}
	}

	head := f.blocks[blockOffset]
	head.WriteAt(buffer.Next(blockSize-bytesOffset), bytesOffset)
	left -= blockSize - bytesOffset

	// required block count to write rest of the data
	required := left / blockSize
	if left%blockSize != 0 {
		required++
	}

	// append additional blocks to file if needed or rewrite existing
	if len(f.blocks) <= blockOffset+required {
		var diff = blockOffset + required - len(f.blocks)

		for i := 0; i <= diff; i++ {
			block := &Block{}
			block.Write(buffer.Next(blockSize))
			f.blocks = append(f.blocks, block)
		}
	} else {
		for i := 1; i <= required; i++ {
			f.blocks[blockOffset+i].Write(buffer.Next(blockSize))
		}
	}

	f.modtime = time.Now()
	return len(p), nil
}

// Read - read all file data
func (f *File) Read() []byte {
	var buffer = new(bytes.Buffer)
	for _, block := range f.blocks {
		buffer.Write(block.Read())
	}
	return buffer.Bytes()
}

// ReadAt - read file data with offset
func (f *File) ReadAt(p []byte, off int) (int, error) {
	var buffer = new(bytes.Buffer)

	blockOffset := off / blockSize
	bytesOffset := off % blockSize

	if len(f.blocks) <= blockOffset {
		return 0, ErrOffsetRange
	}

	head, err := f.blocks[blockOffset].ReadAt(bytesOffset)
	if err != nil {
		return 0, ErrReadBytes
	}
	buffer.Write(head)

	for i := blockOffset + 1; i < len(f.blocks); i++ {
		buffer.Write(f.blocks[i].Read())
	}

	return buffer.Read(p)
}

// Truncate - change file size
func (f *File) Truncate(size int) error {
	blockCount := size / blockSize
	if size%blockSize != 0 {
		blockCount++
	}
	defer func() { f.modtime = time.Now() }()

	if len(f.blocks) > blockCount {
		f.blocks = f.blocks[:blockCount]
		return nil
	}

	var diff = blockCount - len(f.blocks)
	for i := 0; i <= diff; i++ {
		f.blocks = append(f.blocks, &Block{})
	}

	f.modtime = time.Now()
	return nil
}

// Name of the file
func (f *File) Name() string {
	return f.name
}

// IsDir checks if file is directory
func (f *File) IsDir() bool {
	return f.dir
}

// Size in bytes
func (f *File) Size() int {
	return len(f.blocks) * blockSize
}

// StatString -
func (f *File) StatString() string {
	template := "  File: %s\n"
	template += "  Size: %d Blocks: %d "
	if f.IsDir() {
		template += "directory\n"
	} else {
		template += "regular file\n"
	}
	template += "Access: (%s) Links: %d\n"
	template += "Modify: %s\n"

	return fmt.Sprintf(template, f.Name(), f.Size(), len(f.blocks), f.mode, len(f.links), f.modtime.String())
}

// MarshalJSON -
func (f *File) MarshalJSON() ([]byte, error) {
	var data = f.Read()
	var memdir []string

	if f.IsDir() && f.memDir != nil {
		for _, file := range f.memDir.Files() {
			memdir = append(memdir, filepath.Base(file.Name()))
		}
	}

	return json.Marshal(map[string]interface{}{
		"name":    f.name,
		"id":      f.id,
		"links":   f.links,
		"data":    string(data),
		"size":    len(f.blocks) * blockSize,
		"blocks":  len(f.blocks),
		"dir":     f.dir,
		"memdir":  memdir,
		"modtime": f.modtime,
	})
}
