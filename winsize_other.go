//go:build !linux && !darwin && !freebsd && !dragonfly && !netbsd && !openbsd && !solaris
// +build !linux,!darwin,!freebsd,!dragonfly,!netbsd,!openbsd,!solaris

package pty

// Winsize is used to set the size of a terminal.
type Winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}
