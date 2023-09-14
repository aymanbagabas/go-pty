//go:build windows
// +build windows

package main

import (
	"os"

	"github.com/aymanbagabas/go-pty"
)

func notifySizeChanges(chan os.Signal) {}

func handlePtySize(p pty.Pty, _ chan os.Signal) {
	// TODO
}

func initSizeChange(chan os.Signal) {}
