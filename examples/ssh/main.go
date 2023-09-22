package main

import (
	"io"
	"log"

	"github.com/aymanbagabas/go-pty"
	"github.com/charmbracelet/ssh"
)

func main() {
	ssh.Handle(func(s ssh.Session) {
		ptyReq, winCh, isPty := s.Pty()
		if isPty {
			pseudo, err := pty.New()
			if err != nil {
				log.Println(err)
				return
			}

			defer pseudo.Close()
			w, h := ptyReq.Window.Width, ptyReq.Window.Height
			if ptyReq.Modes != nil {
				if err := pty.ApplyTerminalModes(int(pseudo.Fd()), w, h, ptyReq.Modes); err != nil {
					log.Println(err)
					return
				}
			}
			if err := pseudo.Resize(w, h); err != nil {
				log.Println(err)
				return
			}

			cmd := pseudo.Command("bash")
			cmd.Env = append(cmd.Env, "TERM="+ptyReq.Term, "SSH_TTY="+pseudo.Name())

			if err := cmd.Start(); err != nil {
				log.Print(err)
				return
			}

			go func() {
				for win := range winCh {
					pseudo.Resize(win.Height, win.Width)
				}
			}()

			go io.Copy(pseudo, s) // stdin
			go io.Copy(s, pseudo) // stdout

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
