package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
)

type handler func([]string) error

type command struct {
	name    string
	handler handler
	argc    int
}

// Run the command
func (c *command) Run(args []string) {
	if len(args) < c.argc {
		infolog.Printf("%s takes %d arg (s)\n", c.name, c.argc)
		return
	}

	err := c.handler(args)
	if err != nil {
		errlog.Println(err)
	}
}

// Babbler - filesystem's talker
type Babbler struct {
	fspath   string
	mounted  *MemFS
	commands map[string]*command
}

// Babble - create new bubbler
func Babble() *Babbler {
	b := &Babbler{
		mounted:  &MemFS{},
		commands: make(map[string]*command),
	}

	b.Command("help", 0, func(args []string) error {
		for _, cmd := range b.commands {
			fmt.Printf("%s: \n", cmd.name)
			fmt.Print(color.HiMagentaString("\targc %d\n", cmd.argc))
		}
		return nil
	})
	return b
}

// Command registers command handler
func (b *Babbler) Command(name string, argc int, h handler) {
	b.commands[name] = &command{
		name:    name,
		argc:    argc,
		handler: h,
	}
}

// Exec the command
func (b *Babbler) Exec(input string) {
	parts := strings.Fields(input)

	name := parts[0]
	args := parts[1:]

	cmd, ok := b.commands[name]
	if !ok {
		fmt.Println("unknown command")
		return
	}

	if name != "mount" && name != "help" && (b.mounted == nil || len(b.mounted.mount) < 1) {
		errlog.Println("no filesystem mounted")
		return
	}

	cmd.Run(args)
}

// Run - run the cli
func (b *Babbler) Run() {
	cyan.Print("$ ")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()

		if input == "" {
			b.line()
			continue
		}

		b.Exec(input)
		b.line()
	}

	if err := scanner.Err(); err != nil {
		log.Fatalln(err)
	}
}

func (b *Babbler) line() {
	if b.mounted != nil && b.mounted.current != nil {
		blue.Print(b.mounted.Pwd())
		cyan.Print(" $ ")
	} else {
		cyan.Print("$ ")
	}
}
