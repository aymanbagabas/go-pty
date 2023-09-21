package main

import (
	"io"
	"log"
	"os"

	"github.com/aymanbagabas/go-pty"
)

func main() {
	ptmx, tty, err := pty.Open()
	if err != nil {
		log.Fatalf("failed to open pty: %s", err)
	}

	defer ptmx.Close()
	defer tty.Close()

	c := pty.Command("grep", "--color=auto", "bar")
	c.Stdin = tty
	c.Stdout = tty
	c.Stderr = tty

	if err := c.Start(); err != nil {
		log.Fatalf("failed to start: %s", err)
	}

	go func() {
		ptmx.Write([]byte("foo\n"))
		ptmx.Write([]byte("bar\n"))
		ptmx.Write([]byte("baz\n"))
		ptmx.Write([]byte{4}) // EOT
	}()
	go io.Copy(os.Stdout, ptmx)

	if err := c.Wait(); err != nil {
		panic(err)
	}
}
