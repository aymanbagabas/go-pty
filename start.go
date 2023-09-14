package pty

import (
	"context"
	"os"
	"os/exec"
)

// StartOption represents a configuration option passed to Start.
type StartOption func(*startOptions)

type startOptions struct {
	ptyOpts []Option
	ptyCb   PTYCallback
}

// WithPTYOption applies the given options to the underlying PTY.
func WithPTYOption(opts ...Option) StartOption {
	return func(o *startOptions) {
		o.ptyOpts = append(o.ptyOpts, opts...)
	}
}

// PTYCallback is a function that is called with the Cmd and the PTY before the
// command is started.  This allows the caller to modify the PTY and Cmd before
// the command is started.
type PTYCallback func(PTY, *Cmd) error

// WithPTYCallback allows the caller to modify the Cmd before it is started.
func WithPTYCallback(fn PTYCallback) StartOption {
	return func(o *startOptions) {
		o.ptyCb = fn
	}
}

// Cmd is a drop-in replacement for exec.Cmd with most of the same API, but
// it exposes the context.Context to our PTY code so that we can still kill the
// process when the Context expires.  This is required because on Windows, we don't
// start the command using the `exec` library, so we have to manage the context
// ourselves.
type Cmd struct {
	Context context.Context
	Path    string
	Args    []string
	Env     []string
	Dir     string
}

func CommandContext(ctx context.Context, name string, arg ...string) *Cmd {
	return &Cmd{
		Context: ctx,
		Path:    name,
		Args:    append([]string{name}, arg...),
		Env:     os.Environ(),
	}
}

func Command(name string, arg ...string) *Cmd {
	return CommandContext(context.Background(), name, arg...)
}

func (c *Cmd) AsExec() *exec.Cmd {
	//nolint: gosec
	execCmd := exec.CommandContext(c.Context, c.Path, c.Args[1:]...)
	execCmd.Dir = c.Dir
	execCmd.Env = c.Env
	return execCmd
}

// Start the command in a TTY.  The calling code must not use cmd after passing it to the PTY, and
// instead rely on the returned Process to manage the command/process.
func Start(cmd *Cmd, opt ...StartOption) (PTYCmd, Process, error) {
	return startPty(cmd, opt...)
}
