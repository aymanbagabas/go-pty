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
	// Create arbitrary command.
	c := pty.Command(`bash`)

	// Start the command with a pty.
	ptmx, proc, err := pty.Start(c)
	if err != nil {
		return err
	}
	// Make sure to close the pty at the end.
	defer func() { _ = ptmx.Close() }() // Best effort.

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
	go func() { _, _ = io.Copy(ptmx.InputWriter(), os.Stdin) }()
	_, _ = io.Copy(os.Stdout, ptmx.OutputReader())

	return proc.Wait()
}

func main() {
	if err := test(); err != nil {
		log.Fatal(err)
	}
}
