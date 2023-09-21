//go:build windows

package ssh

import (
	"fmt"
	"log"

	"golang.org/x/crypto/ssh"
)

func applyTerminalModesToFd(fd uintptr, width int, height int, modes ssh.TerminalModes, logger *log.Logger) error {
	return fmt.Errorf("not implemented")
}
