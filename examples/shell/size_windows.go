//go:build windows
// +build windows

package main

import (
	"os"

	"github.com/aymanbagabas/go-pty"
)

func notifySizeChanges(chan os.Signal) {}

// windows doesn't support SIGWINCH, so we need to poll the terminal size
// periodically.
func handlePtySize(p pty.File, _ chan os.Signal) {
	// for {
	// 	time.Sleep(2 * time.Second)

	// 	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	// 	if err != nil {
	// 		log.Printf("error getting terminal size: %s", err)
	// 		continue
	// 	}

	// 	if err := p.Resize(w, h); err != nil {
	// 		log.Printf("error resizing pty: %s", err)
	// 	}
	// }
}

func initSizeChange(chan os.Signal) {}
