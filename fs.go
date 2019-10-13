package fs

import (
	"io"
	"os"
)

// Filesystem is an abstract filesystem representation
type Filesystem interface {
	Open(name string) (int, error)
	Create(name string) error
	Read(fd, off, size int) (string, error)
	Write(fd, off, size int, data string) (string, error)

	ReadDir(path string) ([]File, error)
	List() []File
	Pwd() string
	Stat(id int) (File, error)
}

// File represents a file with common operations
type File interface {
	ID() uint64
	Truncate(int) error
	ReadAt(p []byte, off int) (int, error)
	WriteAt(p []byte, off int) (int, error)
	os.FileInfo
	io.Closer
}

// FileInfo represents file descriptor interface
type FileInfo interface {
	os.FileInfo
	ID() uint64
}
