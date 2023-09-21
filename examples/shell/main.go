package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/aymanbagabas/go-pty"
	"golang.org/x/term"
)

type PTY interface {
	Resize(w, h int) error
}

func test() error {
	ptmx, tty, err := pty.Open()
	if err != nil {
		return err
	}

	// Create arbitrary command.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	c := exec.CommandContext(ctx, `bash`)
	// c.Stdin = tty
	// c.Stdout = tty
	// c.Stderr = tty

	defer func() {
		_ = ptmx.Close()
		_ = tty.Close()
	}()

	// Start the command with a pty.
	ptmx, err = pty.Start(c)
	if err != nil {
		return err
	}
	// if err := c.Start(); err != nil {
	// 	return err
	// }
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
	go io.Copy(ptmx, os.Stdin)
	go io.Copy(os.Stdout, ptmx)

	return c.Wait()
}

func main() {
	if err := test(); err != nil {
		log.Fatal(err)
	}
}
