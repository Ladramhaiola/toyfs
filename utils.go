package main

import (
	"os"
	"path/filepath"
	"strings"
)

func initfile(fs *Filesystem, name string) *file {
	id := fs.descriptors[0]
	fs.descriptors = fs.descriptors[1:]

	return &file{
		name: name,
		id:   id,
		buf:  new(buffer),
	}
}

func initdir(fs *Filesystem, name string) *file {
	id := fs.descriptors[0]
	fs.descriptors = fs.descriptors[1:]

	return &file{
		name:   name,
		id:     id,
		dir:    true,
		buf:    new(buffer),
		childs: make(map[string]*file),
	}
}

// SplitPath splits path in segments
func SplitPath(path string) []string {
	path = strings.TrimSpace(path)
	path = strings.TrimSuffix(path, "/")
	if path == "" || path == "." {
		return []string{path}
	}

	if len(path) > 0 && !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "./") {
		path = "./" + path
	}

	return strings.Split(path, "/")
}

// search file with given name/path in fs
func (fs *Filesystem) find(name string) (*file, error) {
	if !strings.HasPrefix(name, "/") {
		// convert relative path to absolute
		name = filepath.Join(fs.wd.AbsPath(), name)
	}
	name = filepath.Clean(name)
	parts := SplitPath(name)

	// handle root dir
	if len(parts) == 1 {
		if parts[0] == "" {
			return fs.root, nil
		}
		if parts[0] == "." {
			return fs.wd, nil
		}
	}

	// begin fs graph traverse
	current := fs.root
	if parts[0] == "." {
		current = fs.wd
	}
	parts = parts[1:]

	// traverse graph and search file
	if len(parts) >= 1 {
		for _, part := range parts[:len(parts)-1] {
			if current.childs == nil {
				return nil, os.ErrNotExist
			}

			if entry, ok := current.childs[part]; ok && entry.dir {
				current = entry
			} else {
				return nil, os.ErrNotExist
			}
		}
	}

	name = parts[len(parts)-1]
	f, ok := current.childs[name]
	if !ok {
		return nil, os.ErrNotExist
	}

	return f, nil
}
