package fs

import (
	"io"
	"os"
)

// Filesystem is an abstract filesystem representation
type Filesystem interface {
	Open(name string) (int, error)
	Create(name string) error
	OpenFile(name string) (File, error)

	ReadDir(path string) ([]File, error)
	List() []File
	Pwd() string
	Stat(id int) (FileInfo, error)
}

// File represents a file with common operations
type File interface {
	Name() string
	ID() uint64
	Truncate(int) error
	ReadAt(p []byte, off int) (int, error)
	WriteAt(p []byte, off int) (int, error)
	io.Closer
}

// FileInfo represents file descriptor interface
type FileInfo interface {
	os.FileInfo
	ID() uint64
}
