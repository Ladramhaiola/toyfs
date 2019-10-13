package memfs

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"time"
)

type fproto struct {
	ID      uint64
	Name    string
	Dir     bool
	Size    int64
	ModTime time.Time
	Childs  map[string]fproto
	Parent  string
	Links   []string
	Data    string
}

type fsproto struct {
	Table   map[uint64]string
	Volumes fproto
	Size    uint64
}

func fileToProto(f *File) fproto {
	var childs = make(map[string]fproto)
	if f.IsDir() && len(f.childs) > 0 {
		for name, f := range f.childs {
			childs[name] = fileToProto(f)
		}
	}

	var parent string
	if f.parent != nil {
		parent = f.parent.AbsPath()
	}

	return fproto{
		ID:      f.id,
		Name:    f.name,
		Dir:     f.dir,
		Parent:  parent,
		Size:    f.size,
		ModTime: f.modtime,
		Childs:  childs,
		Links:   f.links,
		Data:    string(f.Read()),
	}
}

func fileFromProto(fs *MemFS, p *fproto) *File {
	f := &File{
		id:      p.ID,
		name:    p.Name,
		dir:     p.Dir,
		size:    p.Size,
		modtime: p.ModTime,
		links:   p.Links,
	}

	if f.dir && len(p.Childs) > 0 {
		f.childs = make(map[string]*File)
		for name, file := range p.Childs {
			f.childs[name] = fileFromProto(fs, &file)
			f.childs[name].parent = f
		}
	}

	f.Write([]byte(p.Data))
	f.fs = fs
	return f
}

// MarshalJSON for saving
func (f *File) MarshalJSON() ([]byte, error) {
	proto := fileToProto(f)
	return json.Marshal(proto)
}

// MarshalJSON for saving
func (fs *MemFS) MarshalJSON() ([]byte, error) {
	return json.Marshal(&fsproto{
		Size:    fs.ids,
		Table:   fs.table,
		Volumes: fileToProto(fs.root),
	})
}

// UnmarshalJSON for loading
func (fs *MemFS) UnmarshalJSON(data []byte) error {
	proto := &fsproto{}
	if err := json.Unmarshal(data, proto); err != nil {
		return err
	}

	fs.ids = proto.Size
	fs.table = proto.Table
	fs.root = fileFromProto(fs, &proto.Volumes)
	fs.wd = fs.root
	fs.opened = make(map[int]*File)
	return nil
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
	defer f.Close()

	fs := &MemFS{}
	err = json.NewDecoder(f).Decode(fs)
	if err != nil {
		return nil, err
	}

	return fs, nil
}
