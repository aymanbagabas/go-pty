package main

import (
	"io"
	"log"

	"github.com/aymanbagabas/go-pty"
	"github.com/charmbracelet/ssh"
)

func main() {
	ssh.Handle(func(s ssh.Session) {
		cmd := pty.Command("bash")
		ptyReq, winCh, isPty := s.Pty()
		if isPty {
			ptmx, tty, err := pty.Open()
			if err != nil {
				panic(err)
			}

			defer ptmx.Close()
			defer tty.Close()

			cmd.Stdin = tty
			cmd.Stdout = tty
			cmd.Stderr = tty
			cmd.Env = append(cmd.Env, "TERM="+ptyReq.Term, "SSH_TTY="+tty.Name())

			if err := cmd.Start(); err != nil {
				log.Print(err)
				return
			}

			go func() {
				for win := range winCh {
					pty.Resize(ptmx, win.Width, win.Height)
				}
			}()

			go io.Copy(ptmx, s) // stdin
			go io.Copy(s, ptmx) // stdout

			if err := cmd.Wait(); err != nil {
				log.Print(err)
				return
			}
		} else {
			io.WriteString(s, "No PTY requested.\n")
			s.Exit(1)
		}
	})

	log.Println("starting ssh server on port 2222...")
	log.Fatal(ssh.ListenAndServe(":2222", nil))
}
