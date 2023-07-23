//go:build darwin || linux
// +build darwin linux

package pty

import (
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

// openpt allocates a new pseudo-terminal by opening the /dev/ptmx device
func openpt() (*os.File, error) {
	fd, err := syscall.Open("/dev/ptmx", unix.O_RDWR|unix.O_NOCTTY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}

	return os.NewFile(uintptr(fd), "/dev/ptmx"), nil
}

// Open opens a pty master and its corresponding slave.
func Open() (File, File, error) {
	m, slvPath, err := open()
	if err != nil {
		return nil, nil, err
	}

	slv, err := os.OpenFile(slvPath, os.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		return nil, nil, err
	}

	return m, slv, nil
}
