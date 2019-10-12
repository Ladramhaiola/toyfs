package main

import (
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
)

func main() {
	Run()
}

// todo: cli, Link Unlink
