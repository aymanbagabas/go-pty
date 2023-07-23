//go:build linux
// +build linux

package pty

import (
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

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

	slv, err := ptsname(m)
	if err != nil {
		return nil, "", err
	}

	if err := unlockpt(m); err != nil {
		return nil, "", err
	}

	return m, slv, nil
}

func ptsname(f *os.File) (string, error) {
	var n int32
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), syscall.TIOCGPTN, uintptr(unsafe.Pointer(&n)))
	if e != 0 {
		return "", e
	}

	return "/dev/pts/" + strconv.Itoa(int(n)), nil
}

func unlockpt(f *os.File) error {
	var u int32
	// use TIOCSPTLCK with a pointer to zero to clear the lock
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), syscall.TIOCSPTLCK, uintptr(unsafe.Pointer(&u))) //nolint:gosec // Expected unsafe pointer for Syscall call.
	if e != 0 {
		return e
	}

	return nil
}
