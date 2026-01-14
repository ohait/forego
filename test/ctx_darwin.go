//go:build darwin
// +build darwin

package test

import (
	"os"

	"golang.org/x/sys/unix"
)

// isTerminal is true if stdout goes to a console, false if piped
var isTerminal = func() bool {
	_, err := unix.IoctlGetTermios(int(os.Stdout.Fd()), unix.TIOCGETA)
	return err == nil
}()
