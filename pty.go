package pty

import (
	"errors"
	"io"
)

// ErrClosed is returned when the pty is already closed.
var ErrClosed = errors.New("pty: closed")

// File is a generic file descriptor with a name.
// It is used to represent the pty and tty file descriptors.
// On Unix systems, the real type is *os.File
type File interface {
	io.ReadWriteCloser

	// Fd returns the file descriptor number.
	Fd() uintptr

	// Name returns the name of the file.
	// For example /dev/pts/1 or /dev/ttys001.
	// Windows will always return "windows-pty".
	Name() string
}

// Open opens a pty and its corresponding tty.
// On Windows, the pty and tty are the same thing.
func Open() (File, File, error) {
	return open()
}
