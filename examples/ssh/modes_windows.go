//go:build windows
// +build windows

package main

import (
	"log"

	"golang.org/x/crypto/ssh"
)

func applyTerminalModesToFd(fd uintptr, width int, height int, modes ssh.TerminalModes, logger *log.Logger) error {
	return nil
}
