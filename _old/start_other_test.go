//go:build !windows

package pty_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"golang.org/x/xerrors"

	"github.com/aymanbagabas/go-pty"
	"github.com/aymanbagabas/go-pty/ptytest"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestStart(t *testing.T) {
	t.Parallel()
	t.Run("Echo", func(t *testing.T) {
		t.Parallel()
		pty, ps := ptytest.Start(t, pty.Command("echo", "test"))

		pty.ExpectMatch("test")
		err := ps.Wait()
		require.NoError(t, err)
		err = pty.Close()
		require.NoError(t, err)
	})

	t.Run("Kill", func(t *testing.T) {
		t.Parallel()
		pty, ps := ptytest.Start(t, pty.Command("sleep", "30"))
		err := ps.Kill()
		assert.NoError(t, err)
		err = ps.Wait()
		var exitErr *exec.ExitError
		require.True(t, xerrors.As(err, &exitErr))
		assert.NotEqual(t, 0, exitErr.ExitCode())
		err = pty.Close()
		require.NoError(t, err)
	})
}

// these constants/vars are used by Test_Start_copy

const cmdEcho = "echo"

var argEcho = []string{"test"}

// these constants/vars are used by Test_Start_truncate

const (
	countEnd = 1000
	cmdCount = "sh"
)

var argCount = []string{"-c", `
i=0
while [ $i -ne 1000 ]
do
        i=$(($i+1))
        echo "$i"
done
`}

// these constants/vars are used by Test_Start_cancel_context

const cmdSleep = "sleep"

var argSleep = []string{"30"}
