package pty

import (
	"context"
	"errors"
	"io"
)

var (
	// ErrInvalidCommand is returned when the command is invalid.
	ErrInvalidCommand = errors.New("pty: invalid command")

	// ErrUnsupported is returned when the platform is unsupported.
	ErrUnsupported = errors.New("pty: unsupported platform")
)

// New returns a new pseudo-terminal.
func New() (Pty, error) {
	return newPty()
}

// Pty is a pseudo-terminal interface.
type Pty interface {
	io.ReadWriteCloser

	// Name returns the name of the pseudo-terminal.
	// On Windows, this will always be "windows-pty".
	// On Unix, this will return the name of the slave end of the
	// pseudo-terminal TTY.
	Name() string

	// Command returns a command that can be used to start a process
	// attached to the pseudo-terminal.
	Command(name string, args ...string) *Cmd

	// CommandContext returns a command that can be used to start a process
	// attached to the pseudo-terminal.
	CommandContext(ctx context.Context, name string, args ...string) *Cmd

	// Resize resizes the pseudo-terminal.
	Resize(width int, height int) error

	// Fd returns the file descriptor of the pseudo-terminal.
	// On Unix, this will return the file descriptor of the master end.
	// On Windows, this will return the handle of the console.
	Fd() uintptr
}
