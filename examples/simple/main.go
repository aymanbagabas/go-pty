package main

import (
	"io"
	"os"

	"github.com/aymanbagabas/go-pty"
)

func main() {
	c := pty.Command("grep", "--color=auto", "bar")
	cmd, proc, err := pty.Start(c)
	if err != nil {
		panic(err)
	}

	f := cmd.InputWriter()
	go func() {
		f.Write([]byte("foo\n"))
		f.Write([]byte("bar\n"))
		f.Write([]byte("baz\n"))
		f.Write([]byte{4}) // EOT
	}()
	io.Copy(os.Stdout, cmd.OutputReader())
	if err := proc.Wait(); err != nil {
		panic(err)
	}
}
