package main

import (
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/aymanbagabas/go-pty"
)

func main() {
	c := exec.Command("grep", "--color=auto", "bar")
	ptmx, err := pty.Start(c)
	if err != nil {
		log.Fatalf("failed to start: %s", err)
	}

	defer ptmx.Close()
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
