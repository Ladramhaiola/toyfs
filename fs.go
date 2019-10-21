package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
)

// Filesystem represents abstract fs model
type Filesystem struct {
	descriptors []uint64
	root        *file
	wd          *file
	opened      map[int]*file
	table       map[uint64]string
	size        uint64
}

// Create new filesystem
func Create(dcount int) *Filesystem {
	root := &file{
		id:     0,
		dir:    true,
		name:   "/",
		buf:    new(buffer),
		childs: make(map[string]*file),
	}

	fds := make([]uint64, dcount)
	for i := 0; i < dcount; i++ {
		fds[i] = uint64(i + 1)
	}

	return &Filesystem{
		root:        root,
		wd:          root,
		size:        uint64(dcount),
		descriptors: fds,
		opened:      make(map[int]*file),
		table:       make(map[uint64]string),
	}
}

// Create new file
func (fs *Filesystem) Create(name string, dir bool) (f *file, err error) {
	if len(fs.descriptors) == 0 {
		return nil, fmt.Errorf("no free descriptors")
	}

	name = filepath.Clean(name)
	base := filepath.Base(name)

	parent, err := fs.find(filepath.Dir(name))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, &os.PathError{Op: "create", Path: name, Err: err}
		}

		// mkdir parent directory if it doesnt exist
		parent, err = fs.Create(filepath.Dir(name), true)
		if err != nil {
			return nil, err
		}
	}

	if !parent.dir || parent.childs == nil {
		return nil, &os.PathError{Op: "create", Path: name, Err: os.ErrNotExist}
	}

	_, ok := parent.childs[base]
	if ok {
		return nil, &os.PathError{Op: "create", Path: name, Err: os.ErrExist}
	}

	if dir {
		f = initdir(fs, base)
	} else {
		f = initfile(fs, base)
	}
	f.parent = parent

	parent.childs[base] = f
	fs.table[f.id] = f.AbsPath()
	return f, nil
}

// Open file
func (fs *Filesystem) Open(name string) (*file, int, error) {
	f, err := fs.find(name)
	if err != nil {
		return nil, 0, &os.PathError{Op: "open", Path: name, Err: err}
	}

	// if file is symlink -> search & open original file
	origin, ok := f.Symlink()
	if ok {
		f, err = fs.find(origin)
		if err != nil {
			return nil, 0, &os.PathError{Op: "open", Path: name, Err: err}
		}
	}

	if f.dir {
		err = fmt.Errorf("Is a directory")
		return nil, 0, &os.PathError{Op: "open", Path: name, Err: err}
	}

	fd := rand.Intn(100)
	fs.opened[fd] = f
	return f, fd, nil
}

// Close file
func (fs *Filesystem) Close(fd int) {
	delete(fs.opened, fd)
}

// Read file by descriptor
func (fs *Filesystem) Read(fd, off, size int) (string, error) {
	f, ok := fs.opened[fd]
	if !ok {
		return "", fmt.Errorf("file is not opened")
	}

	var data = make([]byte, size)
	_, err := f.ReadAt(data, off)
	return string(data), err
}

// Write data to file by descriptor
func (fs *Filesystem) Write(fd, off, size int, data string) error {
	f, ok := fs.opened[fd]
	if !ok {
		return fmt.Errorf("file is not opened")
	}

	if size < len(data) {
		data = data[:size]
	}

	_, err := f.WriteAt([]byte(data), off)
	return err
}

// Remove file from fs
func (fs *Filesystem) Remove(name string) error {
	f, err := fs.find(name)
	if err != nil {
		return &os.PathError{Op: "rm", Path: name, Err: fmt.Errorf("No such file or directory")}
	}

	if f.dir && len(f.childs) > 0 {
		return &os.PathError{Op: "rm", Path: name, Err: fmt.Errorf("cannot remove nonempty directory")}
	}

	parent, _ := fs.find(filepath.Dir(name))
	*f.buf = buffer{}
	delete(parent.childs, f.name)
	// free descriptor
	fs.descriptors = append(fs.descriptors, f.id)
	return nil
}

// Ls - list files in working directory
func (fs *Filesystem) Ls() []*file {
	var files []*file
	for _, f := range fs.wd.childs {
		files = append(files, f)
	}
	// todo: sort?
	return files
}

// Cd - change directory
func (fs *Filesystem) Cd(path string) error {
	f, err := fs.find(path)
	if err != nil {
		return &os.PathError{Op: "cd", Path: path, Err: err}
	}

	if !f.dir {
		return &os.PathError{Op: "cd", Path: path, Err: fmt.Errorf("%q is not a directory", path)}
	}

	fs.wd = f
	return nil
}

// Pwd - working dir path
func (fs *Filesystem) Pwd() string {
	return fs.wd.AbsPath()
}

// Link file name2 to file name1
func (fs *Filesystem) Link(name1, name2 string) error {
	f, err := fs.find(name1)
	if err != nil {
		return &os.PathError{Op: "link", Path: name1, Err: err}
	}

	if f.dir {
		return &os.PathError{Op: "link", Path: name1, Err: fmt.Errorf("hard link not allowed for directory")}
	}

	l, err := fs.Create(name2, false)
	if err != nil {
		return &os.PathError{Op: "link", Path: name1, Err: err}
	}
	// link to same data buffer (block)
	l.buf = f.buf
	return nil
}

// Symlink file name2 to file name1
func (fs *Filesystem) Symlink(name1, name2 string) error {
	f, err := fs.find(name1)
	if err != nil {
		return &os.PathError{Op: "symlink", Path: name1, Err: err}
	}

	l, err := fs.Create(name2, f.dir)
	if err != nil {
		return &os.PathError{Op: "symlink", Path: name1, Err: err}
	}

	symlink := "sym:" + f.AbsPath()
	_, err = l.WriteAt([]byte(symlink), 0)

	// if parent is directory -> link it's children to symlink
	if f.dir {
		l.childs = f.childs
	}

	return err
}
