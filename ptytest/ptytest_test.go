package ptytest_test

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/aymanbagabas/go-pty/ptytest"
)

func TestPtytest(t *testing.T) {
	t.Parallel()
	t.Run("Echo", func(t *testing.T) {
		t.Parallel()
		pty := ptytest.New(t)
		pty.Output().Write([]byte("write"))
		pty.ExpectMatch("write")
		pty.WriteLine("read")
	})

	t.Run("ReadLine", func(t *testing.T) {
		t.Parallel()
		if runtime.GOOS == "windows" {
			t.Skip("ReadLine is glitchy on windows when it comes to the final line of output it seems")
		}

		ctx, cancel := context.WithTimeout(context.Background(), ptytest.WaitLong)
		t.Cleanup(cancel)
		pty := ptytest.New(t)

		// The PTY expands these to \r\n (even on linux).
		pty.Output().Write([]byte("line 1\nline 2\nline 3\nline 4\nline 5"))
		require.Equal(t, "line 1", pty.ReadLine(ctx))
		require.Equal(t, "line 2", pty.ReadLine(ctx))
		require.Equal(t, "line 3", pty.ReadLine(ctx))
		require.Equal(t, "line 4", pty.ReadLine(ctx))
		require.Equal(t, "line 5", pty.ExpectMatch("5"))
	})

	// See https://github.com/coder/coder/issues/2122 for the motivation
	// behind this test.
	t.Run("Ptytest should not hang when output is not consumed", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name          string
			output        string
			isPlatformBug bool
		}{
			{name: "1024 is safe (does not exceed macOS buffer)", output: strings.Repeat(".", 1024)},
			{name: "1025 exceeds macOS buffer (must not hang)", output: strings.Repeat(".", 1025)},
			{name: "10241 large output", output: strings.Repeat(".", 10241)}, // 1024 * 10 + 1
		}
		for _, tt := range tests {
			tt := tt
			// nolint:paralleltest // Avoid parallel test to more easily identify the issue.
			t.Run(tt.name, func(t *testing.T) {
				inv := &cobra.Command{
					Use: "test",
					RunE: func(cmd *cobra.Command, _ []string) error {
						_, err := fmt.Fprint(cmd.OutOrStdout(), tt.output)
						return err
					},
				}
				pty := ptytest.New(t)
				pty.Attach(inv)
				err := inv.Execute()
				require.NoError(t, err)
			})
		}
	})
}
