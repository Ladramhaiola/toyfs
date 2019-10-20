package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
)

func (f *file) MarshalJSON() ([]byte, error) {
	parent := f.parent.AbsPath()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return json.Marshal(map[string]interface{}{
		"id":     f.id,
		"name":   f.name,
		"parent": parent,
		"data":   string(data),
		"dir":    f.dir,
		"childs": f.childs,
		"size":   f.Size(),
	})
}

// MarshalJSON dumps fs to single json
func (fs *Filesystem) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"maxdesc":     fs.size,
		"root":        fs.root.childs,
		"table":       fs.table,
		"descriptors": fs.descriptors,
	})
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
