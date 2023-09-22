//go:build !linux && !darwin && !freebsd && !dragonfly && !netbsd && !openbsd && !solaris && !windows
// +build !linux,!darwin,!freebsd,!dragonfly,!netbsd,!openbsd,!solaris,!windows

package pty

import (
	"golang.org/x/crypto/ssh"
)

func applyTerminalModesToFd(fd int, width int, height int, modes ssh.TerminalModes) error {
	// TODO
	return nil
}
