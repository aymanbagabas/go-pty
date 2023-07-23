package pty

import (
	"bytes"
	"io"
	"os/exec"
	"sync"
	"testing"
)

func TestConsolePty(t *testing.T) {
	master, slave, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	defer master.Close()
	defer slave.Close()

	iteration := 10

	var (
		b  bytes.Buffer
		wg sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		io.Copy(&b, master) // nolint: errcheck
		wg.Done()
	}()

	for i := 0; i < iteration; i++ {
		cmd := exec.Command("sh", "-c", "printf test")
		cmd.Stdin = slave
		cmd.Stdout = slave
		cmd.Stderr = slave

		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	}
	slave.Close()
	wg.Wait()

	expectedOutput := ""
	for i := 0; i < iteration; i++ {
		expectedOutput += "test"
	}
	if out := b.String(); out != expectedOutput {
		t.Errorf("unexpected output %q", out)
	}
}
