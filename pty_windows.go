//go:build windows
// +build windows

package pty

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	_PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE = 0x20016 // nolint:revive
)

var (
	errClosedConPty = errors.New("pseudo console is closed")
	errNotStarted   = errors.New("process not started")
)

// Install this from github.com/Microsoft/go-winio
// go install github.com/Microsoft/go-winio/tools/mkwinsyscall@latest
//go:generate mkwinsyscall -output zsyscall_windows.go ./*.go

// ConPty is a Windows console pseudo-terminal.
// It uses Windows pseudo console API to create a console that can be used to
// start processes attached to it.
//
// See: https://docs.microsoft.com/en-us/windows/console/creating-a-pseudoconsole-session
type ConPty struct {
	handle          windows.Handle
	inPipe, outPipe *os.File
	mtx             sync.RWMutex
}

var _ Pty = &ConPty{}

func newPty() (Pty, error) {
	ptyIn, inPipeOurs, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create pipes for pseudo console: %w", err)
	}

	outPipeOurs, ptyOut, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create pipes for pseudo console: %w", err)
	}

	var hpc windows.Handle
	coord := windows.Coord{X: 80, Y: 25}
	err = createPseudoConsole(coord, windows.Handle(ptyIn.Fd()), windows.Handle(ptyOut.Fd()), 0, &hpc)
	if err != nil {
		return nil, fmt.Errorf("failed to create pseudo console: %w", err)
	}

	if err := ptyOut.Close(); err != nil {
		return nil, fmt.Errorf("failed to close pseudo console handle: %w", err)
	}
	if err := ptyIn.Close(); err != nil {
		return nil, fmt.Errorf("failed to close pseudo console handle: %w", err)
	}

	return &ConPty{
		handle:  hpc,
		inPipe:  inPipeOurs,
		outPipe: outPipeOurs,
	}, nil
}

// Close implements Pty.
func (p *ConPty) Close() error {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	closePseudoConsole(p.handle)
	return errors.Join(p.inPipe.Close(), p.outPipe.Close())
}

// Command implements Pty.
func (p *ConPty) Command(name string, args ...string) *Cmd {
	c := &Cmd{
		pty:  p,
		Path: name,
		Args: append([]string{name}, args...),
	}
	return c
}

// CommandContext implements Pty.
func (p *ConPty) CommandContext(ctx context.Context, name string, args ...string) *Cmd {
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
func (*ConPty) Name() string {
	return "windows-pty"
}

// Read implements Pty.
func (p *ConPty) Read(b []byte) (n int, err error) {
	return p.outPipe.Read(b)
}

// Resize implements Pty.
func (p *ConPty) Resize(width int, height int) error {
	p.mtx.RLock()
	defer p.mtx.RUnlock()
	if err := resizePseudoConsole(p.handle, windows.Coord{X: int16(height), Y: int16(width)}); err != nil {
		return fmt.Errorf("failed to resize pseudo console: %w", err)
	}
	return nil
}

// Write implements Pty.
func (p *ConPty) Write(b []byte) (n int, err error) {
	return p.inPipe.Write(b)
}

// Fd implements Pty.
func (p *ConPty) Fd() uintptr {
	p.mtx.RLock()
	defer p.mtx.RUnlock()
	return uintptr(p.handle)
}

// updateProcThreadAttribute updates the passed in attribute list to contain the entry necessary for use with
// CreateProcess.
func (p *ConPty) updateProcThreadAttribute(attrList *windows.ProcThreadAttributeListContainer) error {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	if p.handle == 0 {
		return errClosedConPty
	}

	if err := attrList.Update(
		_PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE,
		unsafe.Pointer(p.handle),
		unsafe.Sizeof(p.handle),
	); err != nil {
		return fmt.Errorf("failed to update proc thread attributes for pseudo console: %w", err)
	}

	return nil
}

// createPseudoConsole creates a windows pseudo console.
func createPseudoConsole(size windows.Coord, hInput windows.Handle, hOutput windows.Handle, dwFlags uint32, hpcon *windows.Handle) error {
	// We need this wrapper as the function takes a COORD struct and not a pointer to one, so we need to cast to something beforehand.
	return _createPseudoConsole(*((*uint32)(unsafe.Pointer(&size))), hInput, hOutput, dwFlags, hpcon)
}

// resizePseudoConsole resizes the internal buffers of the pseudo console to the width and height specified in `size`.
func resizePseudoConsole(hpcon windows.Handle, size windows.Coord) error {
	// We need this wrapper as the function takes a COORD struct and not a pointer to one, so we need to cast to something beforehand.
	return _resizePseudoConsole(hpcon, *((*uint32)(unsafe.Pointer(&size))))
}

//sys _createPseudoConsole(size uint32, hInput windows.Handle, hOutput windows.Handle, dwFlags uint32, hpcon *windows.Handle) (hr error) = kernel32.CreatePseudoConsole
//sys _resizePseudoConsole(hPc windows.Handle, size uint32) (hr error) = kernel32.ResizePseudoConsole
//sys closePseudoConsole(hpc windows.Handle) = kernel32.ClosePseudoConsole
