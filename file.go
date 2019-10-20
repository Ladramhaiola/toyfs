package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

type file struct {
	id     uint64
	name   string
	parent *file

	// data buffer
	buf *buffer

	// directory
	dir    bool
	childs map[string]*file
}

func (f *file) Read(p []byte) (n int, err error) {
	return f.buf.Read(p)
}

func (f *file) ReadAt(p []byte, off int) (n int, err error) {
	return f.buf.ReadAt(p, off)
}

func (f *file) Write(p []byte) (n int, err error) {
	return f.buf.Write(p)
}

func (f *file) WriteAt(p []byte, off int) (n int, err error) {
	return f.buf.WriteAt(p, off)
}

func (f *file) Truncate(size int) error {
	return f.buf.Truncate(size)
}

func (f *file) Size() int {
	return f.buf.Len()
}

func (f *file) AbsPath() string {
	if f.parent != nil {
		return filepath.Join(f.parent.AbsPath(), f.name)
	}
	return "/"
}

func (f *file) Symlink() (string, bool) {
	var symflag = make([]byte, 4)
	_, err := f.buf.Read(symflag)
	// symlink file data contains `sym:` flag in 4 first data bytes
	// rest is path to the original file
	if err != nil || string(symflag) != "sym:" {
		return "", false
	}

	link, _ := ioutil.ReadAll(f)
	return string(link[4:]), true
}

func (f *file) String() string {
	var ftype string
	if _, ok := f.Symlink(); ok {
		ftype = "symbolic link"
	} else {
		if f.dir {
			ftype = "directory"
		} else {
			ftype = "regular file"
		}
	}

	return fmt.Sprintf("File: %s id: %5d type: %s size: %d links: %d", f.name, f.id, ftype, f.Size(), 1)
}
