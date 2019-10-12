package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"errors"
)

var lock sync.Mutex

type volume struct {
	Data     string    `json:"data"`
	Dir      bool      `json:"dir"`
	DirNames []string  `json:"memdir"`
	ID       uint64    `json:"id"`
	Links    []string  `json:"links"`
	Linked   string    `json:"linked"`
	Size     int       `json:"size"`
	Blocks   int       `json:"blocks"`
	ModTime  time.Time `json:"modtime"`
	Name     string    `json:"name"`
}

// System in saved description file schema
type System struct {
	BlockSize      int `json:"blocksize"`
	LinkCount      int `json:"linkcount"`
	MaxDescriptors int `json:"descriptors"`
	Volumes        map[string]volume
}

// Marshal is a function that marshals the object into an io.Reader
var Marshal = func(v interface{}) (io.Reader, error) {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

// Save saves a representation of v to the file at path
func Save(path string, v interface{}) error {
	lock.Lock()
	defer lock.Unlock()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	r, err := Marshal(v)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, r)
	return err
}

// Load loads the file at path into
func Load(path string) (*MemFS, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	fsproto := System{}
	err = json.NewDecoder(f).Decode(&fsproto)
	if err != nil {
		return nil, err
	}

	fs := NewMemFs()

	for path, vol := range fsproto.Volumes {
		file := loadfile(vol)
		fs.mount[path] = file
	}

	for path, vol := range fsproto.Volumes {
		if vol.Dir {
			dir, err := loaddir(fs, &vol)
			if err != nil {
				errlog.Println("Loader:", err, path)
				continue
			}

			fs.mount[path].memDir = dir
		}
	}

	fs.count = uint64(len(fs.mount) - 1)
	return fs, err
}

func loadfile(v volume) *File {
	f := &File{
		id:      v.ID,
		name:    v.Name,
		links:   v.Links,
		linked:  v.Linked,
		dir:     v.Dir,
		mode:    os.ModeAppend,
		modtime: v.ModTime,
	}

	f.Write([]byte(v.Data))
	return f
}

func loaddir(fs *MemFS, v *volume) (*DirMap, error) {
	var dirmap = make(DirMap)
	for _, name := range v.DirNames {
		if v.Name != "/" {
			name = v.Name + "/" + name
		}
		f, err := fs.open(name)
		if err != nil {
			return &dirmap, errors.New(err.Error() + " at " + name)
		}

		dirmap.Add(f)
	}
	return &dirmap, nil
}
