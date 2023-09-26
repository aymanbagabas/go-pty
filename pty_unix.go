//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package pty

import (
	"context"
	"errors"
	"os"

	"github.com/creack/pty"
	"golang.org/x/sys/unix"
)

// UnixPty is a POSIX compliant Unix pseudo-terminal.
// See: https://pubs.opengroup.org/onlinepubs/9699919799/
type UnixPty struct {
	master, slave *os.File
	closed        bool
}

var _ Pty = &UnixPty{}

// Close implements Pty.
func (p *UnixPty) Close() error {
	if p.closed {
		return nil
	}
	defer func() {
		p.closed = true
	}()
	return errors.Join(p.master.Close(), p.slave.Close())
}

// Command implements Pty.
func (p *UnixPty) Command(name string, args ...string) *Cmd {
	c := &Cmd{
		pty:  p,
		Path: name,
		Args: append([]string{name}, args...),
	}
	return c
}

// CommandContext implements Pty.
func (p *UnixPty) CommandContext(ctx context.Context, name string, args ...string) *Cmd {
	c := p.Command(name, args...)
	c.ctx = ctx
	return c
}

// Name implements Pty.
func (p *UnixPty) Name() string {
	return p.slave.Name()
}

// Read implements Pty.
func (p *UnixPty) Read(b []byte) (n int, err error) {
	return p.master.Read(b)
}

func (p *UnixPty) Control(f func(fd uintptr)) error {
	conn, err := p.master.SyscallConn()
	if err != nil {
		return err
	}
	return conn.Control(f)
}

// Master returns the pseudo-terminal master end (pty).
func (p *UnixPty) Master() *os.File {
	return p.master
}

// Slave returns the pseudo-terminal slave end (tty).
func (p *UnixPty) Slave() *os.File {
	return p.slave
}

// Winsize represents the terminal window size.
type Winsize = unix.Winsize

// SetWinsize sets the pseudo-terminal window size.
func (p *UnixPty) SetWinsize(ws *Winsize) error {
	var ctrlErr error
	if err := p.Control(func(fd uintptr) {
		ctrlErr = unix.IoctlSetWinsize(int(fd), unix.TIOCSWINSZ, ws)
	}); err != nil {
		return err
	}

	return ctrlErr
}

// Resize implements Pty.
func (p *UnixPty) Resize(width int, height int) error {
	return p.SetWinsize(&Winsize{
		Row: uint16(height),
		Col: uint16(width),
	})
}

// Write implements Pty.
func (p *UnixPty) Write(b []byte) (n int, err error) {
	return p.master.Write(b)
}

// Fd implements Pty.
func (p *UnixPty) Fd() uintptr {
	return p.master.Fd()
}

func newPty() (Pty, error) {
	master, slave, err := pty.Open()
	if err != nil {
		return nil, err
	}

	return &UnixPty{
		master: master,
		slave:  slave,
	}, nil
}
