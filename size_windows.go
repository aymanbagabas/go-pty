//go:build windows
// +build windows

package pty

import (
	"errors"
	"fmt"

	"github.com/aymanbagabas/go-pty/internal/winapi"
	"golang.org/x/sys/windows"
)

func resize(f File, width int, height int) error {
	pty, ok := f.(*conPty)
	if !ok {
		return errors.New("invalid type")
	}

	pty.mtx.RLock()
	defer pty.mtx.RUnlock()

	if pty.handle == 0 {
		return ErrClosed
	}

	coord := windows.Coord{X: int16(width), Y: int16(height)}
	if err := winapi.ResizePseudoConsole(pty.handle, coord); err != nil {
		return fmt.Errorf("failed to resize pseudo console: %w", err)
	}

	return nil
}
