//go:build !windows
// +build !windows

package pty

import (
	"context"
	"errors"
	"os"
	"os/exec"

	"github.com/creack/pty"
	"golang.org/x/sys/unix"
)

// PosixPty is a POSIX compliant pseudo-terminal.
// See: https://pubs.opengroup.org/onlinepubs/9699919799/
type PosixPty struct {
	master, slave *os.File
	closed        bool
}

var _ Pty = &PosixPty{}

// Close implements Pty.
func (p *PosixPty) Close() error {
	if p.closed {
		return nil
	}
	defer func() {
		p.closed = true
	}()
	return errors.Join(p.master.Close(), p.slave.Close())
}

// Command implements Pty.
func (p *PosixPty) Command(name string, args ...string) *Cmd {
	cmd := exec.Command(name, args...)
	c := &Cmd{
		pty:  p,
		sys:  cmd,
		Path: name,
		Args: append([]string{name}, args...),
	}
	c.sys = cmd
	return c
}

// CommandContext implements Pty.
func (p *PosixPty) CommandContext(ctx context.Context, name string, args ...string) *Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	c := p.Command(name, args...)
	c.ctx = ctx
	c.Cancel = func() error {
		return cmd.Cancel()
	}
	return c
}

// Name implements Pty.
func (p *PosixPty) Name() string {
	return p.slave.Name()
}

// Read implements Pty.
func (p *PosixPty) Read(b []byte) (n int, err error) {
	return p.master.Read(b)
}

func (p *PosixPty) Control(f func(fd uintptr)) error {
	conn, err := p.master.SyscallConn()
	if err != nil {
		return err
	}
	return conn.Control(f)
}

// Master returns the pseudo-terminal master end (pty).
func (p *PosixPty) Master() *os.File {
	return p.master
}

// Slave returns the pseudo-terminal slave end (tty).
func (p *PosixPty) Slave() *os.File {
	return p.slave
}

// Winsize represents the terminal window size.
type Winsize = unix.Winsize

// SetWinsize sets the pseudo-terminal window size.
func (p *PosixPty) SetWinsize(ws *Winsize) error {
	var ctrlErr error
	if err := p.Control(func(fd uintptr) {
		ctrlErr = unix.IoctlSetWinsize(int(fd), unix.TIOCSWINSZ, ws)
	}); err != nil {
		return err
	}

	return ctrlErr
}

// Resize implements Pty.
func (p *PosixPty) Resize(width int, height int) error {
	return p.SetWinsize(&Winsize{
		Row: uint16(height),
		Col: uint16(width),
	})
}

// Write implements Pty.
func (p *PosixPty) Write(b []byte) (n int, err error) {
	return p.master.Write(b)
}

// Fd implements Pty.
func (p *PosixPty) Fd() uintptr {
	return p.master.Fd()
}

func newPty() (Pty, error) {
	master, slave, err := pty.Open()
	if err != nil {
		return nil, err
	}

	return &PosixPty{
		master: master,
		slave:  slave,
	}, nil
}
