//go:build !windows
// +build !windows

package pty

import "github.com/creack/pty"

func open() (File, File, error) {
	return pty.Open()
}
