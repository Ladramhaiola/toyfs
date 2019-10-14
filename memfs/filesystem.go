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
		if !os.IsNotExist(err) {
			return &os.PathError{Op: "mkdir", Path: name, Err: err}
		}

		// create parent directory if it doesn't exist
		if err := fs.Mkdir(filepath.Dir(name)); err != nil {
			return err
		}
		_, parent, _ = fs.file(filepath.Dir(name))
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

	if parent.childs == nil {
		parent.childs = make(map[string]*File)
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
		if !os.IsNotExist(err) {
			return &os.PathError{Op: "create", Path: name, Err: err}
		}

		// create parent directory if it doesn't exist
		if err := fs.Mkdir(filepath.Dir(name)); err != nil {
			return err
		}
		parent, _, _ = fs.file(name)
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
	return nil
}

// Pwd - get working directory
func (fs *MemFS) Pwd() string {
	return fs.wd.AbsPath()
}

// Link name2 to name1
func (fs *MemFS) Link(name1, name2 string) error {
	name1 = filepath.Clean(name1)
	name2 = filepath.Clean(name2)
	_, f, err := fs.file(name1)
	if err != nil {
		return &os.PathError{Op: "cat", Path: name1, Err: err}
	}
	if f == nil || f.dir {
		return &os.PathError{Op: "cat", Path: name1, Err: os.ErrNotExist}
	}

	if err := fs.Create(name2); err != nil {
		return err
	}
	f.links = append(f.links, name2)

	_, link, _ := fs.file(name2)
	link.data = f.data
	link.linked = f.AbsPath()
	return nil
}

// Unlink file
func (fs *MemFS) Unlink(name string) error {
	name = filepath.Clean(name)
	p, f, err := fs.file(name)
	if err != nil {
		return &os.PathError{Op: "cat", Path: name, Err: err}
	}
	if f == nil || f.dir {
		return &os.PathError{Op: "cat", Path: name, Err: os.ErrNotExist}
	}

	_, parent, err := fs.file(f.linked)
	if err == nil {
		for i, link := range parent.links {
			if link == f.name {
				if len(parent.links)+1 > i {
					parent.links = append(parent.links[:i], parent.links[i+1:]...)
				} else {
					parent.links = parent.links[:i]
				}
			}
		}
	}

	delete(p.childs, f.name)
	return nil
}

// Remove file
func (fs *MemFS) Remove(name string) error {
	name = filepath.Clean(name)
	parent, f, err := fs.file(name)
	if err != nil {
		return &os.PathError{Op: "cat", Path: name, Err: err}
	}
	if f == nil || f.dir {
		return &os.PathError{Op: "cat", Path: name, Err: os.ErrNotExist}
	}

	delete(parent.childs, f.name)
	return nil
}

// RemoveDir -
func (fs *MemFS) RemoveDir(name string) error {
	name = filepath.Clean(name)
	parent, f, err := fs.file(name)
	if err != nil {
		return &os.PathError{Op: "cat", Path: name, Err: err}
	}
	if f == nil || !f.dir {
		return &os.PathError{Op: "cat", Path: name, Err: os.ErrNotExist}
	}
	if len(f.childs) > 0 {
		return fmt.Errorf("directory is not empty")
	}

	delete(parent.childs, f.name)
	return nil
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
