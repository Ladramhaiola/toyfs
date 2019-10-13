package main

import (
	"fmt"
	"fs/memfs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

var (
	red    = color.New(color.FgHiRed)
	yellow = color.New(color.FgYellow)
	green  = color.New(color.FgGreen)
	cyan   = color.New(color.FgHiCyan)
	blue   = color.New(color.FgBlue)
)

func main() {
	b := Babble()

	b.Command("ls", 0, func(args []string) error {
		for _, f := range b.mounted.List() {
			if f.IsDir() {
				cyan.Printf("%4d %s\n", f.ID(), f.Name())
			} else {
				fmt.Printf("%4d %s\n", f.ID(), f.Name())
			}
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

	b.Command("unmount", 0, func(args []string) error {
		defer func() { b.mounted = nil }()
		return memfs.Save(b.fspath, b.mounted)
	})

	b.Command("mount", 1, func(args []string) error {
		fs, err := memfs.Load(args[0])
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}

			fs = memfs.Create()
			if err := memfs.Save(args[0], fs); err != nil {
				return err
			}
		}
		b.mounted = fs
		b.fspath = args[0]
		return nil
	})

	b.Command("close", 1, func(args []string) error {
		fd, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		return b.mounted.Close(fd)
	})

	b.Command("read", 3, func(args []string) error {
		fd, err := strconv.Atoi(args[0])
		if err != nil {
			return nil
		}
		off, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}
		size, err := strconv.Atoi(args[2])
		if err != nil {
			return err
		}

		data, err := b.mounted.Read(fd, off, size)
		if err != nil {
			return err
		}
		green.Println(data)
		return nil
	})

	b.Command("write", 4, func(args []string) error {
		fd, err := strconv.Atoi(args[0])
		if err != nil {
			return nil
		}
		off, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}
		size, err := strconv.Atoi(args[2])
		if err != nil {
			return err
		}

		// todo: bug - not rewriting existing blocks
		info, err := b.mounted.Write(fd, off, size, strings.Join(args[3:], " "))
		if err != nil {
			return err
		}
		green.Println(info)
		return nil
	})

	// todo: fix truncate
	b.Command("truncate", 2, func(args []string) error {
		size, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}

		return b.mounted.Truncate(args[0], size)
	})

	b.Command("cat", 1, func(args []string) error {
		data, err := b.mounted.Cat(args[0])
		if err != nil {
			return err
		}
		fmt.Println(data)
		return nil
	})

	b.Run()
}
