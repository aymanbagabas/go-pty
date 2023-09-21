//go:build !windows
// +build !windows

package pty

import (
	"errors"
	"os/exec"
	"syscall"
)

func (c *Cmd) start() (err error) {
	cmd := c.asExec()
	if cmd.Stdin == nil || cmd.Stdout == nil || cmd.Stderr == nil {
		pty, tty, err := open()
		if err != nil {
			return err
		}

		defer func() {
			_ = tty.Close()
		}()

		if cmd.Stdin == nil {
			cmd.Stdin = tty
		}
		if cmd.Stdout == nil {
			cmd.Stdout = tty
		}
		if cmd.Stderr == nil {
			cmd.Stderr = tty
		}

		defer func() {
			if err != nil {
				_ = pty.Close()
			}
		}()
	}

	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid:  true,
			Setctty: true,
		}
	}

	startErr := cmd.Start()
	c.sys = cmd
	c.Process = cmd.Process
	return startErr
}

func (c *Cmd) close() error {
	return nil
}

func (c *Cmd) wait() error {
	cmd, ok := c.sys.(*exec.Cmd)
	if !ok {
		return errors.New("invalid type")
	}

	defer func() {
		c.donec <- struct{}{}
	}()

	err := cmd.Wait()
	c.ProcessState = cmd.ProcessState
	return err
}
