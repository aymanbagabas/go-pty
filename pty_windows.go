//go:build windows
// +build windows

package pty

import (
	"fmt"
	"os"
	"sync"
	"unsafe"

	"github.com/creack/pty"
	"golang.org/x/sys/windows"
)

type conPty struct {
	handle windows.Handle

	// These are the Pty I/O ends of the pseudoconsole pipes.
	inputWrite, outputRead *os.File
	mtx                    sync.RWMutex
}

func open() (*conPty, *conPty, error) {
	// We use the CreatePseudoConsole API which was introduced in build 17763
	vsn := windows.RtlGetVersion()
	if vsn.MajorVersion < 10 ||
		vsn.BuildNumber < 17763 {
		return nil, nil, pty.ErrUnsupported
	}

	// Create a pipe for the input and output of the pseudoconsole
	// The
	// See: https://learn.microsoft.com/en-us/windows/console/creating-a-pseudoconsole-session
	//
	inputRead, inputWrite, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	outputRead, outputWrite, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	var handle windows.Handle
	consoleSize := windows.Coord{X: 80, Y: 25}
	if err := windows.CreatePseudoConsole(consoleSize, windows.Handle(inputRead.Fd()), windows.Handle(outputWrite.Fd()), 0, &handle); err != nil {
		return nil, nil, err
	}

	// Now that we created a pty, we can close the tty inputRead and outputWrite
	// channels.
	// See: https://learn.microsoft.com/en-us/windows/console/creating-a-pseudoconsole-session
	if err := inputRead.Close(); err != nil {
		return nil, nil, err
	}
	if err := outputWrite.Close(); err != nil {
		return nil, nil, err
	}

	pty := &conPty{
		handle:     handle,
		inputWrite: inputWrite,
		outputRead: outputRead,
	}

	return pty, pty, nil
}

func (p *conPty) Name() string {
	// See: https://github.com/PowerShell/openssh-portable/blob/latestw_all/contrib/win32/win32compat/win32_sshpty.c#L36
	return "windows-pty"
}

func (p *conPty) Fd() uintptr {
	p.mtx.RLock()
	defer p.mtx.RUnlock()
	return uintptr(p.handle)
}

func (p *conPty) Read(data []byte) (int, error) {
	return p.outputRead.Read(data)
}

func (p *conPty) Write(data []byte) (int, error) {
	return p.inputWrite.Write(data)
}

func (p *conPty) Close() error {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if p.handle == 0 {
		return nil
	}

	// Close the pty and set the handle to 0 to indicate that it's closed.
	windows.ClosePseudoConsole(p.handle)
	p.handle = 0

	// Close the side of the pipes that we own.
	if err := p.inputWrite.Close(); err != nil {
		return err
	}
	if err := p.outputRead.Close(); err != nil {
		return err
	}

	return nil
}

// UpdateProcThreadAttribute updates the passed in attribute list to contain the entry necessary for use with
// CreateProcess.
func (c *conPty) UpdateProcThreadAttribute(attrList *windows.ProcThreadAttributeListContainer) error {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	if c.handle == 0 {
		return ErrClosed
	}

	if err := attrList.Update(
		windows.PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE,
		unsafe.Pointer(c.handle),
		unsafe.Sizeof(c.handle),
	); err != nil {
		return fmt.Errorf("failed to update proc thread attributes for pseudo console: %w", err)
	}

	return nil
}
