package ssh

import (
	"log"

	"golang.org/x/crypto/ssh"
)

// ApplyTerminalModes
// request to the given fd.
//
// This is based on code from Tailscale's tailssh package:
// https://github.com/tailscale/tailscale/blob/main/ssh/tailssh/incubator.go
func ApplyTerminalModes(fd uintptr, width int, height int, modes ssh.TerminalModes, logger *log.Logger) error {
	return applyTerminalModesToFd(fd, height, width, modes, logger)
}
