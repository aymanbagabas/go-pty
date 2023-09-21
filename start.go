package pty

import (
	"os/exec"
	"syscall"
)

func Start(c *exec.Cmd) (File, error) {
	pty, _, err := open()
	if err != nil {
		return nil, err
	}

	c.SysProcAttr = &syscall.SysProcAttr{
		PseudoConsole: syscall.Handle(pty.handle),
	}

	if err := c.Start(); err != nil {
		_ = pty.Close()
		return nil, err
	}

	return pty, nil
}
