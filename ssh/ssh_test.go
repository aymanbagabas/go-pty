package ssh_test

import (
	"io"
	"log"
	"runtime"
	"testing"

	"github.com/aymanbagabas/go-pty"
	"github.com/aymanbagabas/go-pty/ptytest"
	sshPty "github.com/aymanbagabas/go-pty/ssh"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestSSH(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}
	t.Run("SSH_TTY", func(t *testing.T) {
		t.Parallel()
		h, w := 24, 80
		modes := ssh.TerminalModes{
			ssh.ECHO:   1,
			ssh.ICANON: 1,
		}
		logger := log.New(io.Discard, "", 0)
		pty, ps := ptytest.Start(t,
			pty.Command("env"),
			pty.WithPTYOption(pty.WithSize(w, h)),
			pty.WithPTYCallback(func(p pty.PTY, c *pty.Cmd) error {
				c.Env = append(c.Env, "SSH_TTY="+p.Name())
				if c, ok := p.(pty.Controllable); ok {
					if err := c.ControlTTY(func(fd uintptr) error {
						return sshPty.ApplyTerminalModes(fd, w, h, modes, logger)
					}); err != nil {
						return err
					}
				}
				return nil
			}),
		)
		pty.ExpectMatch("SSH_TTY=/dev/")
		err := ps.Wait()
		require.NoError(t, err)
		err = pty.Close()
		require.NoError(t, err)
	})
}
