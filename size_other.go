//go:build !windows
// +build !windows

package pty

import (
	"errors"
	"os"

	"github.com/creack/pty"
)

func resize(f File, width int, height int) error {
	file, ok := f.(*os.File)
	if !ok {
		return errors.New("invalid type")
	}

	return pty.Setsize(file, &pty.Winsize{
		Cols: uint16(width),
		Rows: uint16(height),
	})
}
