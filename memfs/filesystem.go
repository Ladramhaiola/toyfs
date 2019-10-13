package memfs

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	vfs "fs"
)

// MemFS - in-memory filesystem
type MemFS struct {
	root   *File
	wd     *File
	ids    uint64
	table  map[uint64]string
	opened map[int]*File
}

// Create a new MemFS
func Create() *MemFS {
	root := &File{
		name: "/",
		dir:  true,
		id:   0,
	}
	return &MemFS{
		root:   root,
		wd:     root,
		table:  make(map[uint64]string),
		opened: make(map[int]*File),
	}
}

// Mkdir creates a new directory
func (fs *MemFS) Mkdir(name string) error {
	name = filepath.Clean(name)
	base := filepath.Base(name)
	parent, f, err := fs.file(name)
	if err != nil {
		return &os.PathError{Op: "mkdir", Path: name, Err: err}
	}
	if f != nil {
		return &os.PathError{Op: "mkdir", Path: name, Err: fmt.Errorf("directory %q already exists", name)}
	}

	f = &File{
		name:    base,
		id:      atomic.AddUint64(&fs.ids, 1),
		dir:     true,
		mode:    os.ModeTemporary,
		parent:  parent,
		modtime: time.Now(),
		fs:      fs,
	}

	parent.childs[base] = f
	fs.table[f.id] = f.AbsPath()
	return nil
}

// Stat - filestats
func (fs *MemFS) Stat(id int) (vfs.File, error) {
	path, ok := fs.table[uint64(id)]
	if !ok {
		return nil, fmt.Errorf("file with id %d doesn't exist", id)
	}

	_, f, err := fs.file(path)
	return f, err
}

// List file names inside current directory
func (fs *MemFS) List() []vfs.File {
	files := make([]vfs.File, 0, len(fs.wd.childs))
	for _, f := range fs.wd.childs {
		files = append(files, f)
	}
	return files
}

// Create new file
func (fs *MemFS) Create(name string) error {
	name = filepath.Clean(name)
	base := filepath.Base(name)

	parent, f, err := fs.file(name)
	if err != nil {
		return &os.PathError{Op: "create", Path: name, Err: err}
	}

	if f != nil {
		return &os.PathError{Op: "create", Path: name, Err: os.ErrExist}
	}

	f = &File{
		name:    base,
		id:      atomic.AddUint64(&fs.ids, 1),
		dir:     false,
		mode:    os.ModeAppend,
		parent:  parent,
		modtime: time.Now(),
		fs:      fs,
	}

	parent.childs[base] = f
	fs.table[f.id] = f.AbsPath()
	f.modtime = time.Now()
	return nil
}

// Open opens file by its name
func (fs *MemFS) Open(name string) (int, error) {
	name = filepath.Clean(name)
	base := filepath.Base(name)

	_, f, err := fs.file(name)
	if err != nil {
		return 0, &os.PathError{Op: "open", Path: name, Err: err}
	}

	if f == nil {
		return 0, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
	}
	if f.dir {
		return 0, &os.PathError{Op: "open", Path: name, Err: fmt.Errorf("%q is a directory", base)}
	}

	fd := rand.Intn(1000)
	if _, ok := fs.opened[fd]; ok {
		return 0, &os.PathError{Op: "open", Path: name, Err: fmt.Errorf("%q is already opened", name)}
	}
	fs.opened[fd] = f

	return fd, nil
}

// Close the file
func (fs *MemFS) Close(fd int) error {
	delete(fs.opened, fd)
	return nil
}

// Read specified size with offset
func (fs *MemFS) Read(fd, off, size int) (string, error) {
	f, ok := fs.opened[fd]
	if !ok {
		return "", fmt.Errorf("file isn't opened")
	}

	var data = make([]byte, size)
	if _, err := f.ReadAt(data, off); err != nil {
		return "", err
	}
	return string(data), nil
}

// Write data with specified size and offset
func (fs *MemFS) Write(fd, off, size int, data string) (string, error) {
	f, ok := fs.opened[fd]
	if !ok {
		return "", fmt.Errorf("file isn't opened")
	}

	if len(data) > size {
		data = data[:size]
	}
	n, err := f.WriteAt([]byte(data), off)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d bytes written to file", n), nil
}

// Truncate file size
func (fs *MemFS) Truncate(name string, size int) error {
	name = filepath.Clean(name)
	_, f, err := fs.file(name)
	if err != nil {
		return &os.PathError{Op: "readdir", Path: name, Err: err}
	}
	if f == nil || f.dir {
		return &os.PathError{Op: "readdir", Path: name, Err: os.ErrNotExist}
	}

	return f.Truncate(size)
}

// ReadDir reads the directory and returns list of files
func (fs *MemFS) ReadDir(path string) ([]vfs.File, error) {
	path = filepath.Clean(path)
	_, f, err := fs.file(path)
	if err != nil {
		return nil, &os.PathError{Op: "readdir", Path: path, Err: err}
	}
	if f == nil || !f.dir {
		return nil, &os.PathError{Op: "readdir", Path: path, Err: fmt.Errorf("not directory")}
	}

	files := make([]vfs.File, 0, len(f.childs))
	for _, file := range f.childs {
		files = append(files, file)
	}

	// todo: sort by name | id
	return files, nil
}

// Cd change directory
func (fs *MemFS) Cd(path string) error {
	_, f, err := fs.file(filepath.Clean(path))
	if err != nil {
		return &os.PathError{Op: "cd", Path: path, Err: err}
	}
	if f == nil || !f.dir {
		return &os.PathError{Op: "cd", Path: path, Err: fmt.Errorf("not a directory")}
	}

	fs.wd = f
	fmt.Printf("DEBUG cd %+v\n", fs.wd)
	return nil
}

// Pwd - get working directory
func (fs *MemFS) Pwd() string {
	return fs.wd.AbsPath()
}

// Cat - print file data
func (fs *MemFS) Cat(name string) (string, error) {
	_, f, err := fs.file(name)
	if err != nil {
		return "", &os.PathError{Op: "cat", Path: name, Err: err}
	}
	if f == nil || f.dir {
		return "", &os.PathError{Op: "cat", Path: name, Err: os.ErrNotExist}
	}

	return string(f.Read()), nil
}

// find file and its parent in filesystem
func (fs *MemFS) file(path string) (*File, *File, error) {
	if strings.HasPrefix(path, "/") { // if path is absolute
		path = filepath.Clean(path)
	} else { // convert relative path to absolute
		path = filepath.Clean(filepath.Join(fs.wd.AbsPath(), path))
	}
	segs := SplitPath(path)

	// handle root directory
	if len(segs) == 1 {
		if segs[0] == "" {
			return nil, fs.root, nil
		}
		if segs[0] == "." {
			return fs.wd.parent, fs.wd, nil
		}
	}

	// determine root to traverse
	parent := fs.root
	if segs[0] == "." {
		parent = fs.wd
	}
	segs = segs[1:]

	// further directories
	if len(segs) > 1 {
		for _, seg := range segs[:len(segs)-1] {
			if parent.childs == nil {
				return nil, nil, os.ErrNotExist
			}
			if entry, ok := parent.childs[seg]; ok && entry.dir {
				parent = entry
			} else {
				return nil, nil, os.ErrNotExist
			}
		}
	}

	lastSeg := segs[len(segs)-1]
	if parent.childs != nil {
		if node, ok := parent.childs[lastSeg]; ok {
			return parent, node, nil
		}
	} else {
		parent.childs = make(map[string]*File)
	}

	return parent, nil, nil
}
