package pty

import (
	"context"
	"os"
	"syscall"
)

// Cmd is a command that can be started attached to a pseudo-terminal.
// This is similar to the API of exec.Cmd. The main difference is that
// the command is started attached to a pseudo-terminal.
// This is required as we cannot use exec.Cmd directly on Windows due to
// limitation of starting a process attached to a pseudo-terminal.
// See: https://github.com/golang/go/issues/62708
type Cmd struct {
	ctx context.Context
	pty Pty
	sys interface{}

	// Path is the path of the command to run.
	Path string

	// Args holds command line arguments, including the command as Args[0].
	Args []string

	// Env specifies the environment of the process.
	// If Env is nil, the new process uses the current process's environment.
	Env []string

	// Dir specifies the working directory of the command.
	// If Dir is the empty string, the current directory is used.
	Dir string

	// SysProcAttr holds optional, operating system-specific attributes.
	SysProcAttr *syscall.SysProcAttr

	// Process is the underlying process, once started.
	Process *os.Process

	// ProcessState contains information about an exited process.
	// If the process was started successfully, Wait or Run will populate this
	// field when the command completes.
	ProcessState *os.ProcessState

	// Cancel is called when the command is canceled.
	Cancel func() error
}

// Start starts the specified command attached to the pseudo-terminal.
func (c *Cmd) Start() error {
	return c.start()
}

// Wait waits for the command to exit.
func (c *Cmd) Wait() error {
	return c.wait()
}

// Run runs the command and waits for it to complete.
func (c *Cmd) Run() error {
	if err := c.Start(); err != nil {
		return err
	}
	return c.Wait()
}
