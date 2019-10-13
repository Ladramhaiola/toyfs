package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/fatih/color"
)

// MemFS - inmemory filesystem
type MemFS struct {
	count   uint64
	mount   map[string]*File
	opened  map[int]*File
	current *File
}

// NewMemFs - init new in-memory filesystem
func NewMemFs() *MemFS {
	root := CreateDir("/")

	return &MemFS{
		current: root,
		mount:   map[string]*File{"/": root},
		opened:  map[int]*File{},
	}
}

// Filestat by file file id
func (m *MemFS) Filestat(id int) {
	uid := uint64(id)

	var name string
	for n, file := range m.mount {
		if file.id == uid {
			name = n
			break
		}
	}

	file, err := m.open(name)
	if err != nil {
		errlog.Println(err)
		return
	}

	fmt.Println(file.StatString())
}

// Ls lists all files in the root directory
// todo: impl list from current dir
func (m *MemFS) Ls() {

	if m.current.memDir == nil {
		return
	}

	template := "%4d %s\n"
	fmt.Println("total", m.current.memDir.Len())

	for _, file := range m.current.memDir.Files() {
		name := strings.TrimLeft(file.Name(), "/")
		name = filepath.Base(name)

		if file.IsDir() {
			color.Cyan(template, file.id, name)
		} else {
			fmt.Printf(template, file.id, name)
		}
	}
}

// Pwd print current working directory name
func (m *MemFS) Pwd() string {
	path := m.current.Name()
	if path != "/" {
		path = "/" + path
	}
	return path
}

// Cd change directory
func (m *MemFS) Cd(path string) {
	f, err := m.open(path)
	if err != nil {
		fmt.Printf("can't move to %s\n", path)
		return
	}

	if !f.IsDir() {
		errlog.Printf("%s is not directory\n", path)
		return
	}
	m.current = f
}

// Create new file
func (m *MemFS) Create(name string) *File {
	name = m.normPath(name)
	// name = strings.TrimLeft(name, "/")
	file := CreateFile(name)
	m.register(file)
	return file
}

// Open file
func (m *MemFS) Open(name string) (int, error) {
	file, err := m.open(name)
	if err != nil {
		return 0, err
	}

	fd := rand.Int()
	file.closed = false
	m.opened[fd] = file
	return fd, nil
}

// Close by file descriptor
func (m *MemFS) Close(fd int) {
	file, ok := m.opened[fd]
	if !ok {
		infolog.Println("file already closed")
		return
	}

	file.closed = true
	delete(m.opened, fd)
	fmt.Printf("File: %s closed\n", file.Name())
}

// ReadAt find file by descriptor and read with offset
func (m *MemFS) ReadAt(fd, off, size int) {
	file, ok := m.opened[fd]
	if !ok {
		errlog.Printf("file %d isn't opened\n", fd)
		return
	}

	data := make([]byte, size)
	if _, err := file.ReadAt(data, off); err != nil {
		errlog.Println(err)
		return
	}

	fmt.Println(string(data))
}

// WriteAt given data with offset
func (m *MemFS) WriteAt(fd, off int, data string) {
	file, ok := m.opened[fd]
	if !ok {
		errlog.Printf("file %d isn't opened\n", fd)
		return
	}

	n, err := file.WriteAt([]byte(data), off)
	if err != nil {
		errlog.Println(err)
		return
	}

	oklog.Printf("%d bytes written to file\n", n)
}

// Link file2 to file1
func (m *MemFS) Link(name2, name1 string) {
	file, ok := m.mount[name1]
	if !ok {
		errlog.Printf("file %s doesn't exist", name1)
		return
	}

	file.links = append(file.links, name2)
	m.registerLink(file, name2)
}

// Unlink file
func (m *MemFS) Unlink(name string) {
	file, err := m.open(name)
	if err != nil {
		errlog.Println(err)
		return
	}

	parent, err := m.open(file.linked)
	if err == nil {
		for i, link := range parent.links {
			if link == name {
				if i < len(parent.links)+1 {
					parent.links = append(parent.links[:i], parent.links[i+1:]...)
				} else {
					parent.links = parent.links[:i]
				}
			}
		}
	}

	if err := m.unregister(name); err != nil {
		errlog.Println(err)
	}
}

// Truncate file size
func (m *MemFS) Truncate(name string, size int) {
	file, err := m.open(name)
	if err != nil {
		errlog.Println(err)
		return
	}

	if err := file.Truncate(size); err != nil {
		errlog.Println(err)
	}
}

// Mkdir -
func (m *MemFS) Mkdir(name string) error {
	name = m.normPath(name)
	f, ok := m.mount[name]
	if ok && !f.IsDir() {
		return &os.PathError{Op: "mkdir", Path: name, Err: os.ErrExist}
	}

	dir := CreateDir(name)
	m.register(dir)
	return nil
}

// Rmdir -
func (m *MemFS) Rmdir(name string) error {
	f, err := m.open(name)
	if err != nil {
		return err
	}

	if !f.IsDir() {
		return fmt.Errorf("%s is not a directory", name)
	}

	if f.memDir != nil || f.memDir.Len() != 0 {
		return fmt.Errorf("%s is not empty", name)
	}

	if err := m.unregister(name); err != nil {
		return &os.PathError{Op: "rmdir", Path: name, Err: err}
	}
	return nil
}

// Remove -
func (m *MemFS) Remove(name string) error {
	file, err := m.open(name)
	if err != nil {
		return &os.PathError{Op: "rm", Path: name, Err: err}
	}

	if file.IsDir() {
		return &os.PathError{Op: "rm", Path: name, Err: errors.New("can't remove directory")}
	}

	if err := m.unregister(name); err != nil {
		return &os.PathError{Op: "rm", Path: name, Err: err}
	}
	return nil
}

func (m *MemFS) findParent(f *File) *File {
	parentDir := filepath.Dir(f.Name())
	file, err := m.open(parentDir)
	if err != nil {
		return nil
	}
	return file
}

func (m *MemFS) register(f *File) {
	parent := m.findParent(f)
	// mkdir parent directory if it doesn't exist
	if parent == nil {
		parentDir := filepath.Dir(f.Name())
		err := m.Mkdir(parentDir)
		if err != nil {
			return
		}

		parent, err = m.open(parentDir)
		if err != nil {
			return
		}
	}

	if parent.memDir == nil {
		parent.dir = true
		parent.memDir = &DirMap{}
	}
	parent.memDir.Add(f)
	m.mount[f.Name()] = f
	SetID(f, atomic.AddUint64(&m.count, 1))
}

func (m *MemFS) registerLink(f *File, as string) {
	l := m.Create(as)
	l.blocks = f.blocks
	l.linked = f.Name()
	parent := m.findParent(l)
	// mkdir parent directory if it doesn't exist
	if parent == nil {
		parentDir := filepath.Dir(l.Name())
		err := m.Mkdir(parentDir)
		if err != nil {
			return
		}

		parent, err = m.open(parentDir)
		if err != nil {
			return
		}
	}

	if parent.memDir == nil {
		parent.dir = true
		parent.memDir = &DirMap{}
	}
	parent.memDir.Add(l)
	m.mount[l.Name()] = l
}

func (m *MemFS) unregister(name string) error {
	file, err := m.open(name)
	if err != nil {
		return err
	}

	parent := m.findParent(file)
	if parent == nil {
		return errors.New("critical error, parent is nil")
	}

	parent.memDir.Remove(file)
	delete(m.mount, name)
	return nil
}

func (m *MemFS) open(name string) (*File, error) {
	name = m.normPath(name)
	file, ok := m.mount[name]
	if !ok {
		fmt.Println("DEBUG open:", name)
		return nil, os.ErrNotExist
	}
	return file, nil
}

func (m *MemFS) normPath(path string) string {
	path = filepath.Clean(path)

	if path == "." || path == ".." {
		return "/"
	}
	return path
}

func (m *MemFS) absPath(path string) string {
	if strings.HasPrefix(path, "/") {
		return strings.TrimLeft(path, "/")
	}
	return strings.TrimLeft(m.current.Name()+"/"+path, "/")
}

// MarshalJSON -
func (m *MemFS) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"blocksize":   blockSize,
		"linkcount":   LinkCount,
		"descriptors": MaxDescriptors,
		"volumes":     m.mount,
	})
}
