package memfs

import (
	"os"
	"path/filepath"
	"strings"
)

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

func (fs *MemFS) parseSymlink(name string) (string, bool) {
	name = filepath.Clean(name)
	_, f, err := fs.file(name)
	if err != nil {
		return "", false
	}

	if f == nil {
		return "", false
	}

	var data = make([]byte, 4)
	if _, err = f.ReadAt(data, 0); err != nil {
		return "", false
	}

	symlink := string(data)
	return symlink[4:], symlink[:4] == "sym:"
}
