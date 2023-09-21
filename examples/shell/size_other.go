//go:build !windows
// +build !windows

package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/aymanbagabas/go-pty"
)

func notifySizeChanges(ch chan os.Signal) {
	signal.Notify(ch, syscall.SIGWINCH)
}

func handlePtySize(p pty.File, ch chan os.Signal) {
	for range ch {
		// w, h, err := term.GetSize(int(os.Stdin.Fd()))
		// if err != nil {
		// 	log.Printf("error resizing pty: %s", err)
		// 	continue
		// }
		// if err := p.Resize(w, h); err != nil {
		// 	log.Printf("error resizing pty: %s", err)
		// }
	}
}

func initSizeChange(ch chan os.Signal) {
	ch <- syscall.SIGWINCH
}
