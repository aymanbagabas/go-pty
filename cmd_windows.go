//go:build windows
// +build windows

package pty

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/windows"
)

type conPtySys struct {
	done   chan error
	cmdErr error
}

func (c *Cmd) start() error {
	pty, ok := c.pty.(*conPty)
	if !ok {
		return ErrInvalidCommand
	}

	c.sys = &conPtySys{
		done: make(chan error, 1),
	}

	pid, proc, err := pty.Spawn(c.Path, c.Args, &syscall.ProcAttr{
		Dir: c.Dir,
		Env: c.Env,
		Sys: c.SysProcAttr,
	})
	if err != nil {
		return err
	}

	// Grab an *os.Process to avoid reinventing the wheel here. The stdlib has great logic around waiting, exit code status/cleanup after a
	// process has been launched.
	c.Process, err = os.FindProcess(pid)
	if err != nil {
		// If we can't find the process via os.FindProcess, terminate the process as that's what we rely on for all further operations on the
		// object.
		if tErr := windows.TerminateProcess(windows.Handle(proc), 1); tErr != nil {
			return fmt.Errorf("failed to terminate process after process not found: %w", tErr)
		}
		return fmt.Errorf("failed to find process after starting: %w", err)
	}

	if c.ctx != nil {
		go c.waitOnContext()
	}

	return nil
}

func (c *Cmd) waitOnContext() {
	sys := c.sys.(*conPtySys)
	select {
	case <-c.ctx.Done():
		sys.cmdErr = c.Cancel()
		if sys.cmdErr == nil {
			sys.cmdErr = c.ctx.Err()
		}
	case err := <-sys.done:
		sys.cmdErr = err
	}
}

func (c *Cmd) wait() (retErr error) {
	if c.Process == nil {
		return errNotStarted
	}
	if c.ProcessState != nil {
		return errors.New("process already waited on")
	}
	defer func() {
		sys := c.sys.(*conPtySys)
		sys.done <- nil
		if retErr == nil {
			retErr = sys.cmdErr
		}
	}()
	c.ProcessState, retErr = c.Process.Wait()
	if retErr != nil {
		return retErr
	}
	return
}
