package main

import (
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/aymanbagabas/go-pty"
	"golang.org/x/term"
)

type PTY interface {
	Resize(w, h int) error
}

func test() error {
	ptmx, err := pty.New()
	if err != nil {
		return err
	}

	defer ptmx.Close()

	c := ptmx.Command(`bash`)
	if err := c.Start(); err != nil {
		return err
	}

	// Handle pty size.
	ch := make(chan os.Signal, 1)
	notifySizeChanges(ch)
	go handlePtySize(ptmx, ch)
	initSizeChange(ch)
	defer func() { signal.Stop(ch); close(ch) }() // Cleanup signals when done.

	// Set stdin in raw mode.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

	// Copy stdin to the pty and the pty to stdout.
	// NOTE: The goroutine will keep reading until the next keystroke before returning.
	go io.Copy(ptmx, os.Stdin)
	go io.Copy(os.Stdout, ptmx)

	return c.Wait()
}

func main() {
	if err := test(); err != nil {
		log.Fatal(err)
	}
}
