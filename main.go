package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

var (
	hicyan = color.New(color.FgHiCyan)
	hiblue = color.New(color.FgHiBlue)
)

func main() {
	b := new(Babbler)
	b.commands = make(map[string]*command)

	b.Handle("mount", 1, func(c Context) error {
		b.mnt = Create(6)
		color.Green("filesystem mounted")
		return nil
	})

	b.Handle("unmount", 0, func(c Context) error {
		savefs("test_vfs", b.mnt)
		b.mnt = nil
		return nil
	})

	b.Handle("stat", 1, filestat)
	b.Handle("ls", 0, ls)

	b.Handle("create", 1, create)
	b.Handle("open", 1, open)
	b.Handle("close", 1, close)
	b.Handle("read", 3, read)
	b.Handle("write", 4, write)
	b.Handle("truncate", 2, truncate)
	b.Handle("rm", 1, rm)

	b.Handle("link", 2, link)
	b.Handle("symlink", 2, symlink)

	b.Handle("mkdir", 1, mkdir)

	b.Handle("cd", 1, cd)
	b.Handle("pwd", 0, func(c Context) error {
		fmt.Println(c.mnt.Pwd())
		return nil
	})

	b.Handle("cat", 1, cat)

	b.Run()
}

func filestat(c Context) error {
	f, err := c.mnt.find(c.args[0])
	if err != nil {
		return err
	}

	fmt.Println(f)
	return nil
}

func ls(c Context) error {
	files := c.mnt.Ls()
	for _, file := range files {
		if file.dir {
			hiblue.Printf("%5d %s\n", file.id, file.name)
		} else {
			fmt.Printf("%5d %s\n", file.id, file.name)
		}
	}
	return nil
}

func create(c Context) error {
	_, err := c.mnt.Create(c.args[0], false)
	return err
}

func open(c Context) error {
	_, fd, err := c.mnt.Open(c.args[0])
	if err != nil {
		return err
	}

	color.Green("file opened with fd: %d", fd)
	return nil
}

func close(c Context) error {
	fd, err := strconv.Atoi(c.args[0])
	if err != nil {
		return err
	}

	c.mnt.Close(fd)
	return nil
}

func read(c Context) error {
	fd, err := strconv.Atoi(c.args[0])
	if err != nil {
		return err
	}
	off, err := strconv.Atoi(c.args[1])
	if err != nil {
		return err
	}
	size, err := strconv.Atoi(c.args[2])
	if err != nil {
		return err
	}

	data, err := c.mnt.Read(fd, off, size)
	if err != nil {
		return err
	}

	fmt.Println(data)
	return err
}

func write(c Context) error {
	fd, err := strconv.Atoi(c.args[0])
	if err != nil {
		return err
	}
	off, err := strconv.Atoi(c.args[1])
	if err != nil {
		return err
	}
	size, err := strconv.Atoi(c.args[2])
	if err != nil {
		return err
	}

	return c.mnt.Write(fd, off, size, strings.Join(c.args[3:], " "))
}

func truncate(c Context) error {
	size, err := strconv.Atoi(c.args[1])
	if err != nil {
		return err
	}

	f, _, err := c.mnt.Open(c.args[0])
	if err != nil {
		return err
	}

	return f.Truncate(size)
}

func rm(c Context) error {
	return c.mnt.Remove(c.args[0])
}

func link(c Context) error {
	return c.mnt.Link(c.args[0], c.args[1])
}

func symlink(c Context) error {
	return c.mnt.Symlink(c.args[0], c.args[1])
}

func mkdir(c Context) error {
	_, err := c.mnt.Create(c.args[0], true)
	return err
}

func cd(c Context) error {
	return c.mnt.Cd(c.args[0])
}

func cat(c Context) error {
	f, _, err := c.mnt.Open(c.args[0])
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
