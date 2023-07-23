//go:build darwin
// +build darwin

package pty

import (
	"bytes"
	"errors"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// unlockpt unlocks the slave pseudoterminal device corresponding to the master pseudoterminal referred to by f.
// unlockpt should be called before opening the slave side of a pty.
func unlockpt(f *os.File) error {
	return unix.IoctlSetPointerInt(int(f.Fd()), unix.TIOCPTYUNLK, 0)
}

const sys_IOCPARM_MASK = 0x1fff

// ptsname retrieves the name of the first available pts for the given master.
func ptsname(f *os.File) (string, error) {
	n := make([]byte, (syscall.TIOCPTYGNAME>>16)&sys_IOCPARM_MASK)
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), syscall.TIOCPTYGNAME, uintptr(unsafe.Pointer(&n[0]))); errno != 0 {
		return "", errno
	}

	end := bytes.IndexByte(n, 0)
	if end < 0 {
		return "", errors.New("TIOCPTYGNAME string not NUL-terminated")
	}

	return string(n[:end]), nil
}

func grantpt(f *os.File) error {
	return unix.IoctlSetPointerInt(int(f.Fd()), unix.TIOCPTYGRANT, 0)
}

func open() (*os.File, string, error) {
	m, err := openpt()
	if err != nil {
		return nil, "", err
	}

	// In case of error after this point, make sure we close the ptmx fd.
	defer func() {
		if err != nil {
			_ = m.Close() // nolint: errcheck
		}
	}()

	if err := grantpt(m); err != nil {
		return nil, "", err
	}

	if err := unlockpt(m); err != nil {
		return nil, "", err
	}

	s, err := ptsname(m)
	if err != nil {
		return nil, "", err
	}

	return m, s, nil
}
