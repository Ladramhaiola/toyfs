package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type fproto struct {
	ID      uint64
	Name    string
	BlockID string
	Dir     bool
	Childs  map[string]fproto
}

type fsproto struct {
	Size        uint64 `json:"maxdesc"`
	Table       map[uint64]string
	Descriptors []uint64
	Blocks      map[string]string
	Root        map[string]fproto
}

func loadfile(parent *file, blocks map[string]*buffer, proto fproto) *file {
	f := &file{
		id:   proto.ID,
		name: proto.Name,
		buf:  blocks[proto.BlockID],
		dir:  proto.Dir,
	}

	if parent != nil {
		f.parent = parent
	}

	if f.dir && proto.Childs != nil {
		f.childs = make(map[string]*file)
		for name, p := range proto.Childs {
			f.childs[name] = loadfile(f, blocks, p)
		}
	}

	return f
}

func (f *file) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"id":      f.id,
		"name":    f.name,
		"blockid": fmt.Sprintf("%p", f.buf),
		"dir":     f.dir,
		"childs":  f.childs,
		"size":    f.Size(),
	})
}

// MarshalJSON dumps fs to single json
func (fs *Filesystem) MarshalJSON() ([]byte, error) {
	blocks := make(map[string]string)

	loadblock(blocks, fs.root)

	return json.Marshal(map[string]interface{}{
		"maxdesc":     fs.size,
		"root":        fs.root.childs,
		"table":       fs.table,
		"descriptors": fs.descriptors,
		"blocks":      blocks,
	})
}

func loadblock(blocks map[string]string, f *file) {
	if f == nil {
		return
	}
	data, _ := ioutil.ReadAll(f)
	blocks[fmt.Sprintf("%p", f.buf)] = string(data)
	if f.dir && f.childs != nil {
		for _, child := range f.childs {
			loadblock(blocks, child)
		}
	}
}

// UnmarshalJSON loads fs from json
// let your enemies be afraid of this function
func (fs *Filesystem) UnmarshalJSON(data []byte) error {
	fsp := &fsproto{}
	if err := json.Unmarshal(data, fsp); err != nil {
		return err
	}

	// load data blocks and init data buffers
	blocks := make(map[string]*buffer)
	if fsp.Blocks != nil {
		for addr, data := range fsp.Blocks {
			buf := new(buffer)
			buf.Write([]byte(data))
			blocks[addr] = buf
		}
	}

	*fs = *Create(int(fsp.Size))

	fs.size = fsp.Size
	fs.table = fsp.Table
	fs.descriptors = fsp.Descriptors

	for name, proto := range fsp.Root {
		fs.root.childs[name] = loadfile(fs.root, blocks, proto)
	}

	for _, f := range fs.root.childs {
		f.parent = fs.root
	}

	for name, f := range fs.wd.childs["terran"].childs {
		fmt.Println(name, f)
	}

	return nil
}

func savefs(path string, v interface{}) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	r, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}

	_, err = io.Copy(f, bytes.NewReader(r))
	return err
}

func loadfs(path string) (*Filesystem, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fs := &Filesystem{}
	if err := json.NewDecoder(f).Decode(fs); err != nil {
		return nil, err
	}

	return fs, nil
}
