package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
)

// Handler of cli command
type Handler func(c Context) error

// Context - babbler context
type Context struct {
	args []string
	mnt  *Filesystem
}

type command struct {
	name    string
	argc    int
	handler Handler
}

func (cmd *command) Run(c Context) {
	if len(c.args) < cmd.argc {
		color.Yellow("%s takes %d arg(s)\n", cmd.name, cmd.argc)
		return
	}

	if err := cmd.handler(c); err != nil {
		color.HiRed("%s\n", err)
	}
}

// Babbler - filesystem cmd interface
type Babbler struct {
	fsname   string
	mnt      *Filesystem
	commands map[string]*command
}

// Handle - register command handler
func (b *Babbler) Handle(name string, argc int, handler Handler) {
	b.commands[name] = &command{
		name:    name,
		argc:    argc,
		handler: handler,
	}
}

// Parse input & exec matching command
func (b *Babbler) Parse(input string) {
	parts := strings.Fields(input)

	name := parts[0]
	args := parts[1:]

	cmd, ok := b.commands[name]
	if !ok {
		fmt.Println("unknown command")
		return
	}

	if name != "mount" && name != "help" && b.mnt == nil {
		color.HiRed("%s\n", "no filesystem mounted")
		return
	}

	ctx := Context{
		args: args,
		mnt:  b.mnt,
	}

	cmd.Run(ctx)
}

func (b *Babbler) line() {
	if b.mnt != nil {
		hicyan.Printf("%s $ ", b.mnt.Pwd())
	} else {
		hicyan.Print("$ ", "")
	}
}

// Run - run the cli
func (b *Babbler) Run() {
	b.line()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()

		if input == "" {
			b.line()
			continue
		}

		b.Parse(input)
		b.line()
	}

	if err := scanner.Err(); err != nil {
		log.Fatalln(err)
	}
}
