package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procCreatePseudoConsole = kernel32.NewProc("CreatePseudoConsole")
	procClosePseudoConsole  = kernel32.NewProc("ClosePseudoConsole")
)

func main() {
	pty, err := NewConPty()
	if err != nil {
		panic(err)
	}

	defer pty.Close()
	cmd := exec.Command("cmd")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		PseudoConsole: syscall.Handle(pty.handle),
	}

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	go func() {
		pty.inPipe.Write([]byte("exit\r\n"))
	}()

	var buf bytes.Buffer
	go io.Copy(&buf, pty.outPipe)

	cmd.Wait()

	println(buf.String())
}

type ConPty struct {
	handle  windows.Handle
	inPipe  *os.File
	outPipe *os.File
}

func (c *ConPty) Close() error {
	ClosePseudoConsole(c.handle)
	if err := c.inPipe.Close(); err != nil {
		return err
	}
	return c.outPipe.Close()
}

// See https://learn.microsoft.com/en-us/windows/console/creating-a-pseudoconsole-session
func NewConPty() (*ConPty, error) {
	inputRead, inputWrite, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	outputRead, outputWrite, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	var handle windows.Handle
	coord := windows.Coord{X: 80, Y: 25}
	err = CreatePseudoConsole(coord, windows.Handle(inputRead.Fd()), windows.Handle(outputWrite.Fd()), 0, &handle)
	if err != nil {
		return nil, err
	}

	if err := outputWrite.Close(); err != nil {
		return nil, err
	}
	if err := inputRead.Close(); err != nil {
		return nil, err
	}

	return &ConPty{
		handle:  handle,
		inPipe:  inputWrite,
		outPipe: outputRead,
	}, nil
}

func CreatePseudoConsole(coord windows.Coord, in windows.Handle, out windows.Handle, flags uint32, pconsole *windows.Handle) (hr error) {
	size := *((*uint32)(unsafe.Pointer(&coord)))
	r0, _, _ := syscall.Syscall6(procCreatePseudoConsole.Addr(), 5, uintptr(size), uintptr(in), uintptr(out), uintptr(flags), uintptr(unsafe.Pointer(pconsole)), 0)
	if r0 != 0 {
		hr = syscall.Errno(r0)
	}
	return
}

func ClosePseudoConsole(console windows.Handle) {
	syscall.Syscall(procClosePseudoConsole.Addr(), 1, uintptr(console), 0, 0)
	return
}
