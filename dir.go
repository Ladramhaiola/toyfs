package main

import (
	"sort"
	"time"
)

// DirMap -
type DirMap map[string]*File

// Len - count of files in directory
func (m DirMap) Len() int { return len(m) }

// Add new file
func (m DirMap) Add(f *File) {
	m[f.name] = f
	f.modtime = time.Now()
}

// Remove file from directory
func (m DirMap) Remove(f *File) {
	delete(m, f.name)
	f.modtime = time.Now()
}

// Files - get list of all files
func (m DirMap) Files() (files []*File) {
	for _, f := range m {
		files = append(files, f)
	}
	sort.Sort(fileSorter(files))
	return files
}

// implement sort.Interface
type fileSorter []*File

func (f fileSorter) Len() int           { return len(f) }
func (f fileSorter) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f fileSorter) Less(i, j int) bool { return f[i].name < f[j].name }

// Names - list all file names in directory
func (m DirMap) Names() (names []string) {
	for name := range m {
		names = append(names, name)
	}
	return
}
