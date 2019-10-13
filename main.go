package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/fatih/color"
)

const (
	blockSize = 10
	// LinkCount - max links for single file
	LinkCount = 6
	// MaxDescriptors - max count of files in system
	MaxDescriptors = 10
)

var (
	errlog  = color.New(color.FgHiRed)
	infolog = color.New(color.FgYellow)
	oklog   = color.New(color.FgGreen)
	cyan    = color.New(color.FgHiCyan)
	blue    = color.New(color.FgHiBlue)
)

func main() {
	babbler := Babble()

	babbler.Command("mount", 1, func(args []string) error {
		// load existing filesystem from path
		fs, err := Load(args[0])
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}

			// create new filesystem with given path
			fs = NewMemFs()
			if err := Save(args[0], fs); err != nil {
				return err
			}
		}

		babbler.mounted = fs
		babbler.fspath = args[0]
		oklog.Println("filesystem successfully mounted")
		return nil
	})

	babbler.Command("unmount", 0, func(args []string) error {
		if err := Save(babbler.fspath, babbler.mounted); err != nil {
			return err
		}

		babbler.mounted = nil
		return nil
	})

	babbler.Command("filestat", 1, func(args []string) error {
		fd, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		babbler.mounted.Filestat(fd)
		return nil
	})

	babbler.Command("ls", 0, func(args []string) error {
		babbler.mounted.Ls()
		return nil
	})

	babbler.Command("create", 1, func(args []string) error {
		babbler.mounted.Create(args[0])
		return nil
	})

	babbler.Command("open", 1, func(args []string) error {
		fd, err := babbler.mounted.Open(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("file %s opened with fd %d\n", args[0], fd)
		return nil
	})

	babbler.Command("close", 1, func(args []string) error {
		fd, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		babbler.mounted.Close(fd)
		return nil
	})

	babbler.Command("read", 3, func(args []string) error {
		fd, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		off, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}
		size, err := strconv.Atoi(args[2])
		if err != nil {
			return err
		}

		babbler.mounted.ReadAt(fd, off, size)
		return nil
	})

	babbler.Command("write", 4, func(args []string) error {
		fd, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		off, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}
		size, err := strconv.Atoi(args[2])
		if err != nil {
			return err
		}

		data := args[3]
		if len(data) > size {
			data = data[:size]
		}

		babbler.mounted.WriteAt(fd, off, data)
		return nil
	})

	babbler.Command("link", 2, func(args []string) error {
		babbler.mounted.Link(args[1], args[0])
		return nil
	})

	babbler.Command("unlink", 1, func(args []string) error {
		babbler.mounted.Unlink(args[0])
		return nil
	})

	babbler.Command("truncate", 2, func(args []string) error {
		size, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}

		babbler.mounted.Truncate(args[0], size)
		return nil
	})

	babbler.Command("rm", 1, func(args []string) error {
		return babbler.mounted.Remove(args[0])
	})

	babbler.Command("rmdir", 1, func(args []string) error {
		return babbler.mounted.Rmdir(args[0])
	})

	babbler.Command("mkdir", 1, func(args []string) error {
		return babbler.mounted.Mkdir(args[0])
	})

	babbler.Command("cd", 1, func(args []string) error {
		babbler.mounted.Cd(args[0])
		return nil
	})

	babbler.Command("pwd", 0, func(args []string) error {
		fmt.Println(babbler.mounted.Pwd())
		return nil
	})

	babbler.Run()
}
