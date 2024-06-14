//go:build windows
// +build windows

package pty

import (
	"context"
	"errors"
	"os"

	"github.com/aymanbagabas/go-pty/conpty"
)

var (
	errNotStarted = errors.New("process not started")
)

// conPty is a Windows console pseudo-terminal.
// It uses Windows pseudo console API to create a console that can be used to
// start processes attached to it.
//
// See: https://docs.microsoft.com/en-us/windows/console/creating-a-pseudoconsole-session
type conPty struct {
	*conpty.ConPty
}

var _ Pty = &conPty{}

func newPty() (ConPty, error) {
	c, err := conpty.New(conpty.DefaultWidth, conpty.DefaultHeight, 0)
	if err != nil {
		return nil, err
	}

	return &conPty{ConPty: c}, nil
}

// Command implements Pty.
func (p *conPty) Command(name string, args ...string) *Cmd {
	c := &Cmd{
		pty:  p,
		Path: name,
		Args: append([]string{name}, args...),
	}
	return c
}

// CommandContext implements Pty.
func (p *conPty) CommandContext(ctx context.Context, name string, args ...string) *Cmd {
	if ctx == nil {
		panic("nil context")
	}
	c := p.Command(name, args...)
	c.ctx = ctx
	c.Cancel = func() error {
		return c.Process.Kill()
	}
	return c
}

// Name implements Pty.
func (*conPty) Name() string {
	return "windows-pty"
}

// Fd implements Pty.
func (p *conPty) Fd() uintptr {
	return uintptr(p.Handle())
}

// InputPipe implements ConPty.
func (p *conPty) InputPipe() *os.File {
	return p.InPipe()
}

// OutputPipe implements ConPty.
func (p *conPty) OutputPipe() *os.File {
	return p.OutPipe()
}
