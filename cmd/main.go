package main

import (
	"fmt"
	"fs/memfs"
	"path/filepath"
	"strconv"

	"github.com/fatih/color"
)

var (
	red    = color.New(color.FgHiRed)
	yellow = color.New(color.FgYellow)
	oklog  = color.New(color.FgGreen)
	cyan   = color.New(color.FgHiCyan)
	blue   = color.New(color.FgBlue)
)

func main() {
	b := Babble()

	b.Command("ls", 0, func(args []string) error {
		for _, f := range b.mounted.List() {
			fmt.Println(f.ID(), f.Name())
		}
		return nil
	})

	b.Command("create", 1, func(args []string) error {
		return b.mounted.Create(args[0])
	})

	b.Command("open", 1, func(args []string) error {
		fd, err := b.mounted.Open(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("file %s opened with fd: %d\n", filepath.Base(args[0]), fd)
		return nil
	})

	b.Command("mkdir", 1, func(args []string) error {
		return b.mounted.Mkdir(args[0])
	})

	b.Command("cd", 1, func(args []string) error {
		return b.mounted.Cd(args[0])
	})

	b.Command("pwd", 0, func(args []string) error {
		fmt.Println(b.mounted.Pwd())
		return nil
	})

	b.Command("filestat", 1, func(args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		info, err := b.mounted.Stat(id)
		if err != nil {
			return nil
		}

		temp := "File: %s Size: %d Mode %s"
		if info.IsDir() {
			temp += " directory\n"
		} else {
			temp += " regular file\n"
		}
		fmt.Printf(temp, info.Name(), info.Size(), info.Mode())
		return nil
	})

	b.Command("unmount", 1, func(args []string) error {
		return memfs.Save(args[0], b.mounted)
	})

	b.Command("mount", 1, func(args []string) error {
		fs, err := memfs.Load(args[0])
		if err != nil {
			return err
		}
		b.mounted = fs
		return nil
	})

	b.Run()
}
