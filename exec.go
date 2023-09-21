package pty

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
)

// Cmd represents a command to be run in a Pty.
// It is a drop-in replacement for exec.Cmd with almost the same API. This is
// required because on Windows, we don't start the command using the os/exec
// package.
// On Unix, it is just a wrapper around exec.Cmd.
type Cmd struct {
	ctx context.Context

	Path string
	Args []string
	Env  []string
	Dir  string

	Stdin  File
	Stdout File
	Stderr File

	SysProcAttr *syscall.SysProcAttr
	Cancel      func() error

	// Process gets set after the command is started.
	Process *os.Process
	// ProcessState gets set after the command exits through Wait.
	ProcessState *os.ProcessState

	// This is used to store Windows specific process information.
	waitCalled bool

	// donec is used to signal that the process has exited.
	donec chan struct{}

	// sys is used to store platform specific implementation.
	sys interface{}
}

// CommandContext returns the Cmd struct to execute the named program with
// the given arguments.
func CommandContext(ctx context.Context, name string, args ...string) *Cmd {
	c := &Cmd{
		ctx:   ctx,
		Path:  name,
		Args:  append([]string{name}, args...),
		donec: make(chan struct{}, 1),
	}
	if ctx != nil && c.Cancel == nil {
		c.Cancel = func() error {
			return c.kill()
		}
	}
	return c
}

// Command returns the Cmd struct to execute the named program with
// the given arguments.
func Command(name string, args ...string) *Cmd {
	return CommandContext(nil, name, args...)
}

var errNotStarted = errors.New("not started")

// Start starts the specified command but does not wait for it to complete.
func (c *Cmd) Start() error {
	return c.start()
}

// Run starts the specified command and waits for it to complete.
func (c *Cmd) Run() error {
	if err := c.Start(); err != nil {
		return err
	}

	return c.Wait()
}

// Wait waits for the command to exit.
func (c *Cmd) Wait() (err error) {
	return c.wait()
}

func (c *Cmd) kill() error {
	if c.Process == nil {
		return errNotStarted
	}

	defer c.close()
	return c.Process.Kill()
}

func (c *Cmd) asExec() *exec.Cmd {
	var cmd *exec.Cmd
	if c.ctx != nil {
		cmd = exec.CommandContext(c.ctx, c.Path, c.Args[1:]...)
	} else {
		cmd = exec.Command(c.Path, c.Args[1:]...)
	}

	cmd.Dir = c.Dir
	cmd.Env = c.Env
	cmd.Stdin = c.Stdin
	cmd.Stdout = c.Stdout
	cmd.Stderr = c.Stderr
	cmd.SysProcAttr = c.SysProcAttr
	if c.Cancel != nil {
		cmd.Cancel = c.Cancel
	}

	return cmd
}
