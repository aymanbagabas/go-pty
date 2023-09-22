//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package pty

import (
	"os/exec"

	"golang.org/x/sys/unix"
)

func (c *Cmd) start() error {
	cmd, ok := c.sys.(*exec.Cmd)
	if !ok {
		return ErrInvalidCommand
	}
	pty, ok := c.pty.(*UnixPty)
	if !ok {
		return ErrInvalidCommand
	}

	cmd.Stdin = pty.slave
	cmd.Stdout = pty.slave
	cmd.Stderr = pty.slave
	cmd.SysProcAttr = &unix.SysProcAttr{
		Setsid:  true,
		Setctty: true,
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	c.Process = cmd.Process
	return nil
}

func (c *Cmd) wait() error {
	cmd, ok := c.sys.(*exec.Cmd)
	if !ok {
		return ErrInvalidCommand
	}
	err := cmd.Wait()
	c.ProcessState = cmd.ProcessState
	return err
}