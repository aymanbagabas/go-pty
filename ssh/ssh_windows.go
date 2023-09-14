//go:build windows

package ssh

import (
	"log"

	"golang.org/x/crypto/ssh"
	"golang.org/x/xerrors"
)

func applyTerminalModesToFd(fd uintptr, width int, height int, modes ssh.TerminalModes, logger *log.Logger) error {
	return xerrors.Errorf("not implemented")
}
