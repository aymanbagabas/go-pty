//go:build windows
// +build windows

package pty

import (
	"golang.org/x/crypto/ssh"
)

func applyTerminalModesToFd(fd int, width int, height int, modes ssh.TerminalModes) error {
	// TODO
	return nil
}
