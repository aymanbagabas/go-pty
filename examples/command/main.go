package main

import (
	"io"
	"log"
	"os"

	"github.com/aymanbagabas/go-pty"
)

func main() {
	pty, err := pty.New()
	if err != nil {
		log.Fatalf("failed to open pty: %s", err)
	}

	defer pty.Close()
	c := pty.Command("grep", "--color=auto", "bar")
	if err := c.Start(); err != nil {
		log.Fatalf("failed to start: %s", err)
	}

	go func() {
		pty.Write([]byte("foo\n"))
		pty.Write([]byte("bar\n"))
		pty.Write([]byte("baz\n"))
		pty.Write([]byte{4}) // EOT
	}()
	go io.Copy(os.Stdout, pty)

	if err := c.Wait(); err != nil {
		panic(err)
	}
}
