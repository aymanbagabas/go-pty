//go:build windows

package ssh

func applyTerminalModesToFd(fd uintptr, width int, height int, modes ssh.TerminalModes, logger *log.Logger) error {
	return xerrors.Errorf("not implemented")
}
