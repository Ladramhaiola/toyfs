package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// Run the CLI
func Run() {
	cyan.Print("$ ")

	var cli = &CLI{}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()

		if input == "" {
			cyan.Print("$ ")
			continue
		}

		parts := strings.Fields(input)
		args := parts[1:]

		switch parts[0] {
		case "ls":
			cli.checkMounted(cli.Ls, args)
		case "mount":
			cli.Mount(args)
		case "unmount":
			cli.checkMounted(cli.Unmount, args)
		case "filestat":
			cli.checkMounted(cli.Filestat, args)
		case "create":
			cli.checkMounted(cli.Create, args)
		case "open":
			cli.checkMounted(cli.Open, args)
		case "read":
			cli.checkMounted(cli.Read, args)
		case "write":
			cli.checkMounted(cli.Write, args)
		case "truncate":
			cli.checkMounted(cli.Truncate, args)
		default:
			fmt.Println("Unknown command")
		}

		cyan.Print("$ ")
	}

	if err := scanner.Err(); err != nil {
		log.Fatalln(err)
	}
}

// CLI -
type CLI struct {
	path string
	fs   *MemFS
}

// Mount command handler
func (c *CLI) Mount(args []string) {
	if len(args) < 1 {
		fmt.Println("mount takes one argument")
		return
	}

	fs, err := Load(args[0])
	if err != nil {
		if !os.IsNotExist(err) {
			errlog.Println(err)
			return
		}

		fs = NewMemFs()
		err = Save(args[0], fs)
		if err != nil {
			errlog.Println(err)
			return
		}
	}

	c.fs = fs
	c.path = args[0]
	oklog.Println("filesystem successfully mounted")
}

// Unmount command handler
func (c *CLI) Unmount(args []string) {
	if err := Save(c.path, c.fs); err != nil {
		errlog.Println(err)
		return
	}

	c.fs = nil
	fmt.Println("filesystem unmounted")
}

// Filestat command handler
func (c *CLI) Filestat(args []string) {
	if len(args) != 1 {
		fmt.Println("filestat takes 1 argument")
		return
	}

	fid, err := strconv.Atoi(args[0])
	if err != nil {
		errlog.Println(err)
		return
	}

	c.fs.Filestat(fid)
}

// Ls command handler
func (c *CLI) Ls(args []string) {
	c.fs.Ls()
}

// Create command handler
func (c *CLI) Create(args []string) {
	if len(args) < 1 {
		fmt.Println("create takes 1 argument")
		return
	}

	c.fs.Create(args[0])
}

// Open command handler
func (c *CLI) Open(args []string) {
	if len(args) < 1 {
		fmt.Println("open takes 1 argument")
		return
	}

	fd, err := c.fs.Open(args[0])
	if err != nil {
		errlog.Println(err)
		return
	}
	fmt.Printf("file %s opened with fd %d\n", args[0], fd)
}

// Close command handler
func (c *CLI) Close(args []string) {
	if len(args) < 1 {
		fmt.Println("close takes 1 argument")
		return
	}

	fd, err := strconv.Atoi(args[0])
	if err != nil {
		errlog.Println(err)
		return
	}

	c.fs.Close(fd)
}

// Read command handler
func (c *CLI) Read(args []string) {
	if len(args) < 3 {
		fmt.Println("read takes 3 argumnets")
		return
	}

	fd, err := strconv.Atoi(args[0])
	if err != nil {
		errlog.Println(err)
		return
	}

	off, err := strconv.Atoi(args[1])
	if err != nil {
		errlog.Println(err)
		return
	}

	size, err := strconv.Atoi(args[2])
	if err != nil {
		errlog.Println(err)
		return
	}

	c.fs.ReadAt(fd, off, size)
}

// Write command handler
func (c *CLI) Write(args []string) {
	if len(args) < 4 {
		fmt.Println("write takes 4 arguments")
		return
	}

	fd, err := strconv.Atoi(args[0])
	if err != nil {
		errlog.Println(err)
		return
	}

	off, err := strconv.Atoi(args[1])
	if err != nil {
		errlog.Println(err)
		return
	}

	size, err := strconv.Atoi(args[2])
	if err != nil {
		errlog.Println(err)
		return
	}

	data := args[3]
	if len(data) > size {
		data = data[:size]
	}

	c.fs.WriteAt(fd, off, data)
}

// Truncate command handler
func (c *CLI) Truncate(args []string) {
	if len(args) < 2 {
		fmt.Println("truncate tekes 2 arguments")
		return
	}

	size, err := strconv.Atoi(args[1])
	if err != nil {
		errlog.Println(err)
		return
	}

	c.fs.Truncate(args[0], size)
}

func (c *CLI) checkMounted(next func([]string), args []string) {
	if c.fs == nil || c.fs.mount == nil {
		errlog.Println("no mounted filesystem")
		return
	}

	next(args)
}
